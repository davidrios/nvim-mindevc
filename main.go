package main

import (
	"os"
	"path/filepath"

	"github.com/davidrios/nvim-mindevc/cmd"
)

func main() {
	cmdToExecute := cmd.RootCmd
	invocationName := filepath.Base(os.Args[0])

	if foundCmd, _, err := cmd.RootCmd.Find([]string{invocationName}); err == nil {
		cmdToExecute = foundCmd
		cmd.RootCmd.RemoveCommand(foundCmd)
		cmdToExecute.Use = invocationName
	}

	if err := cmdToExecute.Execute(); err != nil {
		os.Exit(1)
	}
}
