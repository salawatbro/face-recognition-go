package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"face/config"
	"face/internal/storage"

	"github.com/spf13/cobra"
)

func NewDeleteCmd(cfg *config.Config) *cobra.Command {
	var (
		userID  string
		confirm bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a user from the system",
		Long:  `Delete a user and all their associated face images from the system.`,
		Example: `  face delete --id abc-123
  face delete --id abc-123 --confirm`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cfg, userID, confirm)
		},
	}

	cmd.Flags().StringVar(&userID, "id", "", "user ID to delete (required)")
	cmd.Flags().BoolVarP(&confirm, "confirm", "y", false, "skip confirmation prompt")

	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func runDelete(cfg *config.Config, userID string, confirm bool) error {
	db, err := cfg.GetDatabaseConnection()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	stor, err := storage.NewFileSystemStorage(cfg.FacesDir)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	user, err := db.GetUser(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	fmt.Printf("\nUser to delete:\n")
	fmt.Printf("  ID:    %s\n", user.ID)
	fmt.Printf("  Name:  %s\n", user.Name)
	fmt.Printf("  Faces: %d\n", len(user.Faces))

	if !confirm {
		fmt.Print("\nAre you sure you want to delete this user? (yes/no): ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" && response != "y" {
			fmt.Println("Deletion canceled.")
			return nil
		}
	}

	for _, face := range user.Faces {
		if err := stor.DeleteImage(face.Filename); err != nil {
			fmt.Printf("Warning: failed to delete image %s: %v\n", face.Filename, err)
		}
	}

	if err := db.DeleteUser(userID); err != nil {
		return fmt.Errorf("failed to delete user from database: %w", err)
	}

	fmt.Printf("\n✓ User '%s' deleted successfully\n", user.Name)

	return nil
}
