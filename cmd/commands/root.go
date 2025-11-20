package commands

import (
	"fmt"
	"log"

	"github.com/rayhan889/intern-payroll/internal/application"
	"github.com/rayhan889/intern-payroll/internal/config"
	"github.com/spf13/cobra"
)

var cfgFile string
var argVersionShort bool
var argVersionSemantic bool

var RootCmd = &cobra.Command{
	Use:   "payroll",
	Short: "Intern Payroll monolith Go application",
	Long:  "Intern Payroll monolith Go application to automatically sends bonuses to interns",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the application version",
	Run: func(cmd *cobra.Command, args []string) {
		if argVersionShort {
			fmt.Printf("%s (%s)\n", application.Version, application.BuildHash)
			return
		} else if argVersionSemantic {
			fmt.Printf("%s\n", application.Version)
			return
		} else {
			fmt.Printf("Intern Payroll %s (%s) %s %s\n", application.Version, application.BuildHash, application.BuildDate, application.Platform)
		}
	},
}

func init() {
	// Initialize the configuration
	_, err := config.Load(cfgFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	RootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	RootCmd.PersistentFlags().StringP("env", "e", ".env", "Env file to use")

	RootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&argVersionShort, "short", "s", false, "Show short version")
	versionCmd.Flags().BoolVarP(&argVersionSemantic, "semantic", "S", false, "Show semantic version")
}
