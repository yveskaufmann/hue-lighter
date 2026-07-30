package cli

import (
	"fmt"
	"os"

	"com.github.yveskaufmann/hue-lighter/internal/app"
	"github.com/spf13/cobra"
)

type RootOptions struct {
	// Shutdown, when true, instructs hue-lighter to send a shutdown event
	Shutdown bool
}

func NewCommand() *cobra.Command {
	o := &RootOptions{}

	cmd := &cobra.Command{
		Use:   "hue-lighter",
		Short: "Automatically control Philips Hue lights based on sunrise/sunset and system events",
		Long: "hue-lighter automatically turns your Philips Hue lights on or off based on " +
			"sunrise and sunset times, and turns them off when the machine shuts down.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
	}

	cmd.Flags().BoolVar(&o.Shutdown, "shutdown", false,
		"Send a shutdown event to turn off configured lights and exit, instead of starting the daemon")

	return cmd
}

// Run executes the root command's behavior based on the collected options.
func (o *RootOptions) Run() error {
	appInstance := app.Bootstrap()

	if o.Shutdown {
		if err := appInstance.SendShutdownEvent(); err != nil {
			return fmt.Errorf("failed to send shutdown event: %w", err)
		}
		return nil
	}

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
