package cmd

import (
	"fmt"

	"face/config"
	"face/internal/database"

	"github.com/spf13/cobra"
)

func NewMigrateCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration commands",
		Long:  `Run database migrations to create or update the database schema.`,
	}

	cmd.AddCommand(newMigrateUpCmd(cfg))
	cmd.AddCommand(newMigrateDownCmd(cfg))
	cmd.AddCommand(newMigrateStatusCmd(cfg))

	return cmd
}

func newMigrateUpCmd(cfg *config.Config) *cobra.Command {
	var steps int

	cmd := &cobra.Command{
		Use:   "up",
		Short: "Run pending migrations",
		Long:  `Apply all pending database migrations or a specific number of steps.`,
		Example: `  face migrate up
  face migrate up --steps 1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrateUp(cfg, steps)
		},
	}

	cmd.Flags().IntVarP(&steps, "steps", "n", 0, "number of migrations to run (0 = all)")

	return cmd
}

func newMigrateDownCmd(cfg *config.Config) *cobra.Command {
	var steps int

	cmd := &cobra.Command{
		Use:   "down",
		Short: "Rollback migrations",
		Long:  `Rollback database migrations. By default rolls back one migration.`,
		Example: `  face migrate down
  face migrate down --steps 2
  face migrate down --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")
			return runMigrateDown(cfg, steps, all)
		},
	}

	cmd.Flags().IntVarP(&steps, "steps", "n", 1, "number of migrations to rollback")
	cmd.Flags().Bool("all", false, "rollback all migrations")

	return cmd
}

func newMigrateStatusCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current migration version",
		Long:  `Display the current database migration version.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrateStatus(cfg)
		},
	}
}

func runMigrateUp(cfg *config.Config, steps int) error {
	migrator, err := database.NewMigrator(cfg.DatabaseType, cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	if steps > 0 {
		if err := migrator.Steps(steps); err != nil {
			return err
		}
		fmt.Printf("Applied %d migration(s)\n", steps)
	} else {
		if err := migrator.Up(); err != nil {
			return err
		}
		fmt.Println("All migrations applied successfully")
	}

	return nil
}

func runMigrateDown(cfg *config.Config, steps int, all bool) error {
	migrator, err := database.NewMigrator(cfg.DatabaseType, cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	if all {
		if err := migrator.Down(); err != nil {
			return err
		}
		fmt.Println("All migrations rolled back")
	} else {
		if err := migrator.Steps(-steps); err != nil {
			return err
		}
		fmt.Printf("Rolled back %d migration(s)\n", steps)
	}

	return nil
}

func runMigrateStatus(cfg *config.Config) error {
	migrator, err := database.NewMigrator(cfg.DatabaseType, cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	version, dirty, err := migrator.Version()
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	fmt.Printf("Current version: %d\n", version)
	if dirty {
		fmt.Println("Warning: Database is in dirty state")
	}

	return nil
}
