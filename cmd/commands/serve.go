package commands

import (
	"fmt"
	"os"

	"github.com/rayhan889/neatspace/internal/config"
	"github.com/rayhan889/neatspace/internal/observer/logger"
	"github.com/rayhan889/neatspace/internal/server"
	"github.com/spf13/cobra"
)

func init() {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start application HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Get()

			// Initialize application logger
			logger := logger.SetupLogging(logger.LoggerOpts{
				Level:       cfg.GetSlogLevel(),
				Format:      cfg.Logging.Format,
				NoColor:     cfg.Logging.NoColor,
				Environment: cfg.App.Mode,
			})

			// Initialize HTTP server
			httpAddr := fmt.Sprintf("%s:%d", cfg.App.ServerHost, cfg.App.ServerPort)
			srv := server.NewHTTPServer(httpAddr, logger)
			if err := srv.Run(); err != nil {
				logger.Error("HTTP server exited with error", "err", err)
				os.Exit(1)
			}

			return nil
		},
	}

	RootCmd.AddCommand(serveCmd)
}
