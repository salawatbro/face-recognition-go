package main

import (
	"fmt"
	"os"

	"face/cmd"
	"face/config"

	"github.com/spf13/cobra"
)

var (
	cfg     *config.Config
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "face",
	Short: "Face identification CLI tool",
	Long: `A command-line tool for face detection and identification.
Enroll users with face images and identify people from photos.`,
	Version: "2.0.0",
}

func init() {
	cfg = config.LoadConfig()

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&cfg.DatabasePath, "db", cfg.DatabasePath, "path to database file")
	rootCmd.PersistentFlags().StringVar(&cfg.FacesDir, "faces-dir", cfg.FacesDir, "directory for face images")
	rootCmd.PersistentFlags().Float64Var(&cfg.DefaultThreshold, "threshold", cfg.DefaultThreshold, "matching threshold (0.0-1.0)")

	rootCmd.AddCommand(cmd.NewEnrollCmd(cfg))
	rootCmd.AddCommand(cmd.NewIdentifyCmd(cfg))
	rootCmd.AddCommand(cmd.NewVerifyCmd(cfg))
	rootCmd.AddCommand(cmd.NewListCmd(cfg))
	rootCmd.AddCommand(cmd.NewDeleteCmd(cfg))
	rootCmd.AddCommand(cmd.NewUpdateCmd(cfg))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
