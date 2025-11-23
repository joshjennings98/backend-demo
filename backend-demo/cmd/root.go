package cmd

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/joshjennings98/backend-demo/server/v2/server"
)

var (
	commandFile string
	port        int
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&commandFile, "command", "c", "", "Command file to use the presentation")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 8080, "Port to run server on")

	_ = viper.BindPFlag("command", rootCmd.PersistentFlags().Lookup("command"))
	_ = viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
}

var rootCmd = &cobra.Command{
	Use:   "backend-demo",
	Short: "Demonstrate backend projects",
	Long:  "Demonstrate backend projects with the power of Go and HTMX",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if commandFile == "" {
			err = errors.New("commands file must be provided via -c/-commands")
			return
		}

		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		s, err := server.NewServer(logger, port, commandFile)
		if err != nil {
			return
		}

		err = s.Start(context.Background())
		return
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
