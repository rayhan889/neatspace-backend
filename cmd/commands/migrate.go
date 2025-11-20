package commands

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rayhan889/intern-payroll/internal/config"
	"github.com/rayhan889/intern-payroll/migrations"
	"github.com/spf13/cobra"
)

var forceReset bool
var argUp bool
var argSeed bool

var forceSeed bool

var migrateUpCmd = &cobra.Command{
	Use:   "migrate:up",
	Short: "Apply the latest database migration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()
		migrator := migrations.NewMigrator(cfg.GetDatabaseURL())
		if err := migrator.MigrateUp(cmd.Context()); err != nil {
			log.Fatalf("Failed to apply database migration: %v", err)
		}
		if err := migrator.Close(); err != nil {
			log.Fatalf("Failed to close database connection: %v", err)
		}
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "migrate:status",
	Short: "Show the status of database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()
		migrator := migrations.NewMigrator(cfg.GetDatabaseURL())
		if err := migrator.MigrateStatus(cmd.Context()); err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}
		if err := migrator.Close(); err != nil {
			log.Fatalf("Failed to close database connection: %v", err)
		}
	},
}

var migrateVersionCmd = &cobra.Command{
	Use:   "migrate:version",
	Short: "Show the current database migration version",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()
		migrator := migrations.NewMigrator(cfg.GetDatabaseURL())
		if err := migrator.MigrateVersion(cmd.Context()); err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}
		if err := migrator.Close(); err != nil {
			log.Fatalf("Failed to close database connection: %v", err)
		}
	},
}

var migrateCreateCmd = &cobra.Command{
	Use:   "migrate:create [migration_name]",
	Short: "Create new database migration file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()
		migrationName := args[0]
		migrator := migrations.NewMigrator(cfg.GetDatabaseURL())
		if err := migrator.MigrateCreate(cmd.Context(), migrationName); err != nil {
			log.Fatalf("Failed to create new migration: %v", err)
		}
		if err := migrator.Close(); err != nil {
			log.Fatalf("Failed to close database connection: %v", err)
		}
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "migrate:down [steps]",
	Short: "Rollback the last or N database migrations",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()
		migrator := migrations.NewMigrator(cfg.GetDatabaseURL())
		steps := ""
		if len(args) > 0 {
			steps = args[0]
		}
		if err := migrator.MigrateDown(cmd.Context(), steps); err != nil {
			log.Fatalf("Failed to rollback database migration: %v", err)
		}
		if err := migrator.Close(); err != nil {
			log.Fatalf("Failed to close database connection: %v", err)
		}
	},
}

var migrateSeedCmd = &cobra.Command{
	Use:   "migrate:seed",
	Short: "Seed the database with initial data",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()

		if !forceSeed {
			fmt.Print("Are you sure you want to seed the database? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))
			if input != "y" && input != "yes" {
				fmt.Println("Aborted.")
				return
			}
		}

		// Call SeedInitialData to seed initial data
		migrator := migrations.NewMigrator(cfg.GetDatabaseURL())
		if err := migrator.SeedInitialData(cmd.Context()); err != nil {
			log.Fatalf("Failed to seed initial data: %v", err)
		}

		// Close database connection after seeding
		if err := migrator.Close(); err != nil {
			log.Fatalf("Failed to close database connection: %v", err)
		}

	},
}

var migrateResetCmd = &cobra.Command{
	Use:   "migrate:reset",
	Short: "Rollback all database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()

		if !forceReset {
			fmt.Print("Are you sure you want to rollback all database migrations? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))
			if input != "y" && input != "yes" {
				fmt.Println("Aborted.")
				return
			}
		}

		migratorReset := migrations.NewMigrator(cfg.GetDatabaseURL())
		err := migratorReset.MigrateReset(cmd.Context())
		if err != nil {
			log.Fatalf("Failed to reset database migration: %v", err)
		}
		if err := migratorReset.Close(); err != nil {
			log.Fatalf("Failed to close database connection: %v", err)
		}

		// If seed called but not up, return an error
		if argSeed && !argUp {
			log.Println("Cannot run seeders without running migrations up first, please use --up flag.")
			return
		}

		if argUp {
			migratorUp := migrations.NewMigrator(cfg.GetDatabaseURL())
			if err := migratorUp.MigrateUp(cmd.Context()); err != nil {
				log.Fatalf("Failed to apply database migration: %v", err)
			}
			if err := migratorUp.Close(); err != nil {
				log.Fatalf("Failed to close database connection: %v", err)
			}

			if argSeed {
				seedArgs := make([]string, len(args))
				copy(seedArgs, args)
				if forceReset {
					seedArgs = append(seedArgs, "--force")
				}
				// Set the "force" flag for migrateSeedCmd and check error
				if err := migrateSeedCmd.Flags().Set("force", fmt.Sprintf("%v", forceReset)); err != nil {
					log.Printf("Failed to set force flag for seed command: %v", err)
				}
				migrateSeedCmd.Run(cmd, seedArgs)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(migrateUpCmd)
	RootCmd.AddCommand(migrateStatusCmd)
	RootCmd.AddCommand(migrateVersionCmd)
	RootCmd.AddCommand(migrateCreateCmd)
	RootCmd.AddCommand(migrateDownCmd)

	migrateResetCmd.Flags().BoolVar(&forceReset, "force", false, "Force reset without confirmation")
	migrateResetCmd.Flags().BoolVar(&argUp, "up", false, "Run migrations up after reset")
	migrateResetCmd.Flags().BoolVar(&argSeed, "seed", false, "Run seeders after migration up")
	RootCmd.AddCommand(migrateResetCmd)

	migrateSeedCmd.Flags().BoolVar(&forceSeed, "force", false, "Force seed without confirmation")
	RootCmd.AddCommand(migrateSeedCmd)
}
