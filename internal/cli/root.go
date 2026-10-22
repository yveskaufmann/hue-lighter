package cli

import (
	"fmt"
	"os"

	"com.github.yveskaufmann/hue-lighter/internal/app"
	"github.com/spf13/cobra"
)

type RootOptions struct {
	// NoTray, when true, runs hue-lighter without the tray icon (e.g. headless or in a container)
	NoTray bool
}

func NewCommand() *cobra.Command {
	o := &RootOptions{}

	cmd := &cobra.Command{
		Use:   "hue-lighter",
		Short: "Automatically control Philips Hue lights based on sunrise/sunset and system events",
		Long: "hue-lighter automatically turns your Philips Hue lights on or off based on " +
			"sunrise and sunset times, and turns them off when it is stopped (SIGTERM, SIGINT or SIGHUP), " +
			"e.g. when the machine shuts down.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
	}

	cmd.Flags().BoolVar(&o.NoTray, "no-tray", false,
		"Run without the tray icon, e.g. on headless systems or in containers")

	return cmd
}

// Run executes the root command's behavior based on the collected options.
func (o *RootOptions) Run() error {
	appInstance := app.Bootstrap(!o.NoTray)

	appInstance.Logger().Info("Starting hue-lighter application with PID=", os.Getpid())

	if err := appInstance.Run(); err != nil {
		return fmt.Errorf("unhandled error: %w", err)
	}

	return nil
}

// Execute builds the root command and runs it against os.Args.
func Execute() error {
	return NewCommand().Execute()
}
