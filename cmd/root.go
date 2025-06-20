package cmd

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/davidrios/nvim-mindevc/config"
)

var RootCmd = &cobra.Command{
	Use:   "nvim-mindevc",
	Short: "Setup neovim in a remote environment.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Println("nvim-mindevc", config.VERSION)
			return nil
		}

		return fmt.Errorf("no command or flags provided")
	},
}

var cmdConfig config.ConfigViper
var configFile string
var devcontainerFile string
var verbose bool
var showVersion bool

func init() {
	cobra.OnInitialize(initConfig)

	var subcommand string
	if len(os.Args) > 1 {
		subcommand = filepath.Base(os.Args[1])
	}

	if subcommand != "git" {
		RootCmd.Flags().BoolVar(
			&showVersion,
			"version",
			false,
			"show version and exit")

		RootCmd.PersistentFlags().BoolVarP(
			&verbose,
			"verbose", "v",
			false,
			"show debug messages")

		RootCmd.PersistentFlags().StringVarP(
			&configFile,
			"config", "c",
			"",
			"load settings from config file")

		RootCmd.PersistentFlags().StringVarP(
			&devcontainerFile,
			"devcontainer", "d",
			"",
			"load devcontainer spec from this file")
	}
}

func initConfig() {
	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	var err error
	cmdConfig, err = config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}
