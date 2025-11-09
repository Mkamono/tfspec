package main

import (
	"os"

	"github.com/Mkamono/tfspec/app"
)

func main() {
	app := app.NewTfspecApp()
	rootCmd := app.CreateRootCommand()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
