package cmd

import (
	"fmt"

	"face/config"
	"face/internal/database"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func NewUpdateCmd(cfg *config.Config) *cobra.Command {
	var (
		userID     string
		name       string
		email      string
		phone      string
		addFace    string
		removeFace string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update user information or manage face images",
		Long:  `Update user information such as name, email, phone, or add/remove face images.`,
		Example: `  face update --id abc-123 --email new@example.com
  face update --id abc-123 --add-face photo.jpg
  face update --id abc-123 --remove-face face-uuid`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cfg, userID, name, email, phone, addFace, removeFace)
		},
	}

	cmd.Flags().StringVar(&userID, "id", "", "user ID to update (required)")
	cmd.Flags().StringVar(&name, "name", "", "update user name")
	cmd.Flags().StringVar(&email, "email", "", "update user email")
	cmd.Flags().StringVar(&phone, "phone", "", "update user phone")
	cmd.Flags().StringVar(&addFace, "add-face", "", "add a new face image")
	cmd.Flags().StringVar(&removeFace, "remove-face", "", "remove a face by face ID")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func runUpdate(cfg *config.Config, userID, name, email, phone, addFace, removeFace string) error {
	fs, err := NewFaceSystem(cfg)
	if err != nil {
		return err
	}
	defer fs.Close()

	user, err := fs.DB.GetUser(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	updated := false

	if name != "" {
		user.Name = name
		updated = true
		fmt.Printf("✓ Updated name to: %s\n", name)
	}

	if email != "" {
		user.Email = email
		updated = true
		fmt.Printf("✓ Updated email to: %s\n", email)
	}

	if phone != "" {
		user.Phone = phone
		updated = true
		fmt.Printf("✓ Updated phone to: %s\n", phone)
	}

	if removeFace != "" {
		if err := removeFaceFromUser(fs, userID, removeFace, user); err != nil {
			return err
		}
		updated = true
	}

	if addFace != "" {
		if err := addFaceToUser(fs, userID, addFace); err != nil {
			return err
		}
		updated = true
	}

	if updated && (name != "" || email != "" || phone != "") {
		if err := fs.DB.UpdateUser(user); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
	}

	if !updated {
		fmt.Println("No changes specified. Use --help to see available options.")
	} else {
		fmt.Println("\n✓ Update completed successfully")
	}

	return nil
}

func removeFaceFromUser(fs *FaceSystem, userID, faceID string, user *database.User) error {
	var faceFilename string
	found := false
	for _, face := range user.Faces {
		if face.ID == faceID {
			faceFilename = face.Filename
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("face ID not found")
	}

	if err := fs.DB.RemoveFace(userID, faceID); err != nil {
		return fmt.Errorf("failed to remove face from database: %w", err)
	}

	if err := fs.Storage.DeleteImage(faceFilename); err != nil {
		fmt.Printf("Warning: failed to delete image file: %v\n", err)
	}

	fmt.Printf("✓ Removed face: %s\n", faceID)
	return nil
}

func addFaceToUser(fs *FaceSystem, userID, imagePath string) error {
	fmt.Println("\nAdding new face image...")
	fmt.Println("Detecting face...")

	result, err := fs.ProcessImage(imagePath)
	if err != nil {
		return err
	}

	fmt.Printf("Face detected (quality: %.2f)\n", result.QualityScore)

	if result.QualityScore < 0.3 {
		return fmt.Errorf("quality too low (%.2f), minimum required: 0.30", result.QualityScore)
	}

	faceID := uuid.New().String()
	filename, err := fs.Storage.SaveImage(userID, faceID, result.CroppedFace)
	if err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	faceData := &database.Face{
		ID:           faceID,
		Filename:     filename,
		Embedding:    result.Embedding,
		QualityScore: result.QualityScore,
	}

	if err := fs.DB.AddFace(userID, faceData); err != nil {
		_ = fs.Storage.DeleteImage(filename)
		return fmt.Errorf("failed to add face to database: %w", err)
	}

	fmt.Printf("✓ Face added successfully (ID: %s)\n", faceID)
	return nil
}
