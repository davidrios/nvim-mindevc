package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var outputConfigFile string
var outputForce bool

var showConfigCmd = &cobra.Command{
	Use:   "show-config",
	Short: "Show current configuration values.",
	RunE: func(cmd *cobra.Command, args []string) error {
		yamlData, err := yaml.Marshal(cmdConfig.Viper.AllSettings())
		if err != nil {
			return fmt.Errorf("Failed to marshal config to YAML: %w", err)
		}

		filePath := outputConfigFile
		if filePath == "" {
			os.Stdout.Write(yamlData)
		} else {
			if _, err := os.Stat(filePath); err == nil && !outputForce {
				return fmt.Errorf("Config file, '%s' already exists.", filePath)
			}

			err = os.WriteFile(filePath, yamlData, 0644)
			if err != nil {
				return fmt.Errorf("Failed to write config file to '%s': %w", filePath, err)
			}

			slog.Debug("Configuration written to", "path", filePath)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(showConfigCmd)

	showConfigCmd.Flags().StringVarP(
		&outputConfigFile,
		"output", "o",
		"",
		"save configuration to output file")

	showConfigCmd.Flags().BoolVarP(
		&outputForce,
		"force", "f",
		false,
		"overwrite file if it exists")
}
