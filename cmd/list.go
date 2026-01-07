package cmd

import (
	"encoding/json"
	"fmt"

	"face/config"

	"github.com/spf13/cobra"
)

func NewListCmd(cfg *config.Config) *cobra.Command {
	var (
		formatJSON bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all enrolled users",
		Long:  `Display a list of all users enrolled in the face recognition system.`,
		Example: `  face list
  face list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cfg, formatJSON)
		},
	}

	cmd.Flags().BoolVar(&formatJSON, "json", false, "output in JSON format")

	return cmd
}

func runList(cfg *config.Config, formatJSON bool) error {
	db, err := cfg.GetDatabaseConnection()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	users, err := db.ListUsers()
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	if len(users) == 0 {
		fmt.Println("No users enrolled yet.")
		return nil
	}

	if formatJSON {
		jsonData, err := json.MarshalIndent(users, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(jsonData))
		return nil
	}

	fmt.Printf("\nTotal users: %d\n\n", len(users))

	for i := range users {
		fmt.Printf("[%d] %s\n", i+1, users[i].Name)
		fmt.Printf("    ID:         %s\n", users[i].ID)
		if users[i].Email != "" {
			fmt.Printf("    Email:      %s\n", users[i].Email)
		}
		if users[i].Phone != "" {
			fmt.Printf("    Phone:      %s\n", users[i].Phone)
		}
		fmt.Printf("    Faces:      %d\n", len(users[i].Faces))
		fmt.Printf("    Created:    %s\n", users[i].CreatedAt.Format("2006-01-02 15:04:05"))

		if len(users[i].Metadata) > 0 {
			fmt.Println("    Metadata:")
			for key, value := range users[i].Metadata {
				fmt.Printf("      %s: %v\n", key, value)
			}
		}

		if i < len(users)-1 {
			fmt.Println()
		}
	}

	return nil
}
