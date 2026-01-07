package cmd

import (
	"errors"
	"fmt"
	"log"

	"face/config"
	"face/internal/database"
	"face/internal/face"

	"github.com/spf13/cobra"
)

func NewIdentifyCmd(cfg *config.Config) *cobra.Command {
	var (
		imagePath string
		threshold float64
	)

	cmd := &cobra.Command{
		Use:   "identify",
		Short: "Identify a person from an image",
		Long: `Identify a person by analyzing their face in a provided image.
The system will detect the face, extract embeddings, and match against the database.`,
		Example: `  face identify --image photo.jpg
  face identify --image unknown.jpg --threshold 0.7`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIdentify(cfg, imagePath, threshold)
		},
	}

	cmd.Flags().StringVarP(&imagePath, "image", "i", "", "path to image file (required)")
	cmd.Flags().Float64VarP(&threshold, "threshold", "t", cfg.DefaultThreshold, "matching threshold (0.0-1.0)")
	err := cmd.MarkFlagRequired("image")
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return cmd
}

func runIdentify(cfg *config.Config, imagePath string, threshold float64) error {
	fmt.Println("Initializing face recognition system...")

	fs, err := NewFaceSystem(cfg)
	if err != nil {
		return err
	}
	defer fs.Close()

	matcher := face.NewMatcher(fs.DB)

	fmt.Printf("\nAnalyzing image: %s\n\n", imagePath)
	fmt.Println("Detecting face...")

	result, err := fs.ProcessImage(imagePath)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Face detected (quality: %.2f)\n", result.QualityScore)

	if result.QualityScore < 0.2 {
		fmt.Println("⚠ Warning: Low quality face detected, results may be inaccurate")
	}

	users, err := fs.DB.ListUsers()
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	if len(users) == 0 {
		fmt.Println("\n✗ Database is empty")
		fmt.Println("  Please enroll at least one user first using:")
		fmt.Println("  face enroll --name \"Your Name\" --images \"photo.jpg\"")
		return nil
	}

	fmt.Printf("Matching against %d users in database...\n", len(users))

	allMatches, err := matcher.FindBestMatches(result.Embedding, 5)
	if err != nil {
		return fmt.Errorf("failed to find matches: %w", err)
	}

	if len(allMatches) > 0 {
		fmt.Println("\nTop matches:")
		for i, match := range allMatches {
			fmt.Printf("  %d. %s (%.2f%%)\n", i+1, match.User.Name, match.Confidence*100)
		}
		fmt.Println()
	}

	match, err := matcher.Match(result.Embedding, threshold)
	if err != nil {
		if errors.Is(err, database.ErrNoMatch) {
			fmt.Println("✗ No match found")
			fmt.Printf("  No user matched with confidence >= %.0f%%\n", threshold*100)
			return nil
		}
		return fmt.Errorf("matching failed: %w", err)
	}

	printMatchResult(match)
	return nil
}

func printMatchResult(match *database.MatchResult) {
	fmt.Println("\n✓ Match found!")
	fmt.Println("─────────────────────────────────────")
	fmt.Printf("User ID:     %s\n", match.User.ID)
	fmt.Printf("Name:        %s\n", match.User.Name)
	if match.User.Email != "" {
		fmt.Printf("Email:       %s\n", match.User.Email)
	}
	if match.User.Phone != "" {
		fmt.Printf("Phone:       %s\n", match.User.Phone)
	}
	fmt.Printf("Confidence:  %.2f%%\n", match.Confidence*100)
	fmt.Printf("Face ID:     %s\n", match.FaceID)

	if len(match.User.Metadata) > 0 {
		fmt.Println("\nMetadata:")
		for key, value := range match.User.Metadata {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}
}
