package main

import (
	"fmt"
	"os"

	"face/cmd"
	"face/config"
	"face/internal/database"

	"github.com/spf13/cobra"
)

var (
	cfg     *config.Config
	verbose bool
	dbType  string
)

var rootCmd = &cobra.Command{
	Use:   "face",
	Short: "Face identification CLI tool",
	Long: `A command-line tool for face detection and identification.
Enroll users with face images and identify people from photos.

Supported database backends:
  - sqlite (default): Local file-based database
  - postgres: PostgreSQL server database
  - json: Legacy JSON file database`,
	Version: "2.0.0",
}

func init() {
	cfg = config.LoadConfig()

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&dbType, "db-type", string(cfg.DatabaseType), "database type (sqlite, postgres, json)")
	rootCmd.PersistentFlags().StringVar(&cfg.DatabasePath, "db", cfg.DatabasePath, "database path or connection string")
	rootCmd.PersistentFlags().StringVar(&cfg.FacesDir, "faces-dir", cfg.FacesDir, "directory for face images")
	rootCmd.PersistentFlags().Float64Var(&cfg.DefaultThreshold, "threshold", cfg.DefaultThreshold, "matching threshold (0.0-1.0)")

	// Update config with flag values before each command runs
	cobra.OnInitialize(func() {
		cfg.DatabaseType = database.ParseDatabaseType(dbType)
	})

	rootCmd.AddCommand(cmd.NewEnrollCmd(cfg))
	rootCmd.AddCommand(cmd.NewIdentifyCmd(cfg))
	rootCmd.AddCommand(cmd.NewVerifyCmd(cfg))
	rootCmd.AddCommand(cmd.NewListCmd(cfg))
	rootCmd.AddCommand(cmd.NewDeleteCmd(cfg))
	rootCmd.AddCommand(cmd.NewUpdateCmd(cfg))
	rootCmd.AddCommand(cmd.NewMigrateCmd(cfg))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
