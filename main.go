package main

import (
	"os"

	"github.com/Mkamono/tfspec/app/cmd"
)

func main() {
	app := cmd.NewTfspecApp()
	rootCmd := app.CreateRootCommand()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
