package cmd

import (
	"fmt"

	"face/config"
	"face/internal/face"

	"github.com/spf13/cobra"
)

func NewVerifyCmd(cfg *config.Config) *cobra.Command {
	var (
		userID    string
		imagePath string
		threshold float64
	)

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify if a face image belongs to a specific user",
		Long: `Verify if a given image matches a specific user in the database (1:1 verification).
This is different from identify which searches all users (1:N identification).`,
		Example: `  face verify --user-id abc123 --image photo.jpg
  face verify -u abc123 -i unknown.jpg --threshold 0.7`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVerify(cfg, userID, imagePath, threshold)
		},
	}

	cmd.Flags().StringVarP(&userID, "user-id", "u", "", "user ID to verify against (required)")
	cmd.Flags().StringVarP(&imagePath, "image", "i", "", "path to image file (required)")
	cmd.Flags().Float64VarP(&threshold, "threshold", "t", cfg.DefaultThreshold, "matching threshold (0.0-1.0)")
	_ = cmd.MarkFlagRequired("user-id")
	_ = cmd.MarkFlagRequired("image")

	return cmd
}

func runVerify(cfg *config.Config, userID, imagePath string, threshold float64) error {
	fmt.Println("Initializing face verification system...")

	fs, err := NewFaceSystem(cfg)
	if err != nil {
		return err
	}
	defer fs.Close()

	user, err := fs.DB.GetUser(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	matcher := face.NewMatcher(fs.DB)

	fmt.Printf("\nVerifying image against user: %s\n", user.Name)
	fmt.Printf("User ID: %s\n\n", userID)
	fmt.Println("Detecting face...")

	result, err := fs.ProcessImage(imagePath)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Face detected (quality: %.2f)\n", result.QualityScore)

	if result.QualityScore < 0.2 {
		fmt.Println("⚠ Warning: Low quality face detected, results may be inaccurate")
	}

	matched, confidence, err := matcher.Verify(userID, result.Embedding, threshold)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	fmt.Println("\n─────────────────────────────────────")
	if matched {
		fmt.Println("✓ VERIFIED - Face matches the user!")
		fmt.Printf("Confidence:  %.2f%%\n", confidence*100)
		fmt.Printf("Threshold:   %.2f\n", threshold)
		fmt.Printf("\nUser ID:     %s\n", user.ID)
		fmt.Printf("Name:        %s\n", user.Name)
		if user.Email != "" {
			fmt.Printf("Email:       %s\n", user.Email)
		}
		if user.Phone != "" {
			fmt.Printf("Phone:       %s\n", user.Phone)
		}
	} else {
		fmt.Println("✗ NOT VERIFIED - Face does not match the user")
		fmt.Printf("Confidence:  %.2f%%\n", confidence*100)
		fmt.Printf("Threshold:   %.2f\n", threshold)
		fmt.Printf("\nThe face in the image does not belong to user '%s'\n", user.Name)
	}

	return nil
}
