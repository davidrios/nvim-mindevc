package cmd

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/davidrios/nvim-mindevc/config"
)

var rootCmd = &cobra.Command{
	Use:   "nvim-mindevc",
	Short: "Setup neovim inside devcontainer.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var cmdConfig config.ConfigViper
var configFile string
var devcontainerFile string
var verbose bool

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolVarP(
		&verbose,
		"verbose", "v",
		false,
		"load settings from config file")

	rootCmd.PersistentFlags().StringVarP(
		&configFile,
		"config", "c",
		"",
		"load settings from config file")

	rootCmd.PersistentFlags().StringVarP(
		&devcontainerFile,
		"devcontainer", "d",
		"",
		"load devcontainer spec from this file")

	// GlobalConfig.RuntimeViper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
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
