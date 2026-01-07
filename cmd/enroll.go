package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"face/config"
	"face/internal/database"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func NewEnrollCmd(cfg *config.Config) *cobra.Command {
	var (
		name     string
		email    string
		phone    string
		images   string
		metadata string
	)

	cmd := &cobra.Command{
		Use:   "enroll",
		Short: "Enroll a new user with face images",
		Long: `Enroll a new user by providing their information and one or more face images.
The system will detect faces, extract embeddings, and store them in the database.`,
		Example: `  face enroll --name "John Doe" --email "john@example.com" --images "img1.jpg,img2.jpg"
  face enroll --name "Jane Smith" --images "photo.jpg" --metadata '{"department":"Engineering"}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnroll(cfg, name, email, phone, images, metadata)
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "user name (required)")
	cmd.Flags().StringVarP(&email, "email", "e", "", "user email")
	cmd.Flags().StringVarP(&phone, "phone", "p", "", "user phone number")
	cmd.Flags().StringVarP(&images, "images", "i", "", "comma-separated image paths (required)")
	cmd.Flags().StringVarP(&metadata, "metadata", "m", "", "JSON metadata")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("images")

	return cmd
}

func runEnroll(cfg *config.Config, name, email, phone, imagesStr, metadataStr string) error {
	fmt.Println("Initializing face recognition system...")

	fs, err := NewFaceSystem(cfg)
	if err != nil {
		return err
	}
	defer fs.Close()

	imagePaths := strings.Split(imagesStr, ",")
	for i := range imagePaths {
		imagePaths[i] = strings.TrimSpace(imagePaths[i])
	}

	var metadataMap database.Metadata
	if metadataStr != "" {
		if err := json.Unmarshal([]byte(metadataStr), &metadataMap); err != nil {
			return fmt.Errorf("invalid metadata JSON: %w", err)
		}
	}

	userID := uuid.New().String()
	user := &database.User{
		ID:       userID,
		Name:     name,
		Email:    email,
		Phone:    phone,
		Metadata: metadataMap,
		Faces:    []database.Face{},
	}

	fmt.Printf("\nEnrolling user: %s\n", name)
	fmt.Printf("Processing %d image(s)...\n\n", len(imagePaths))

	for idx, imgPath := range imagePaths {
		fmt.Printf("[%d/%d] Processing %s...\n", idx+1, len(imagePaths), imgPath)

		result, err := fs.ProcessImage(imgPath)
		if err != nil {
			fmt.Printf("  ✗ %v\n", err)
			continue
		}

		fmt.Printf("  • Face detected (quality: %.2f)\n", result.QualityScore)

		if result.QualityScore < 0.3 {
			fmt.Printf("  ✗ Quality too low, skipping\n")
			continue
		}

		faceID := uuid.New().String()
		filename, err := fs.Storage.SaveImage(userID, faceID, result.CroppedFace)
		if err != nil {
			fmt.Printf("  ✗ Failed to save image: %v\n", err)
			continue
		}

		user.Faces = append(user.Faces, database.Face{
			ID:           faceID,
			Filename:     filename,
			Embedding:    database.Embedding(result.Embedding),
			QualityScore: result.QualityScore,
		})
		fmt.Printf("  ✓ Face enrolled successfully\n")
	}

	if len(user.Faces) == 0 {
		return fmt.Errorf("no faces were successfully enrolled")
	}

	if err := fs.DB.CreateUser(user); err != nil {
		for _, faceData := range user.Faces {
			_ = fs.Storage.DeleteImage(faceData.Filename)
		}
		return fmt.Errorf("failed to save user to database: %w", err)
	}

	fmt.Printf("\n✓ User enrolled successfully!\n")
	fmt.Printf("  User ID: %s\n", userID)
	fmt.Printf("  Name: %s\n", name)
	fmt.Printf("  Faces enrolled: %d\n", len(user.Faces))

	return nil
}
