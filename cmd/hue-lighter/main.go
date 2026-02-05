package main

import (
	"os"

	"com.github.yveskaufmann/hue-lighter/internal/app"
)

func main() {
	appInstance := app.Bootstrap()

	for arg := range os.Args {
		{
			if os.Args[arg] == "--shutdown" {
				err := appInstance.SendShutdownEvent()
				if err != nil {
					appInstance.Logger().Fatalf("failed to send shutdown event: %v", err)
				}
				return
			}
		}
	}

	appInstance.Logger().Info("Starting hue-lighter application with PID=", os.Getpid())

	if err := appInstance.Run(); err != nil {
		appInstance.Logger().Fatalf("Unhandled error: %v", err)
	}
}
