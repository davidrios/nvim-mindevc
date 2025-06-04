package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Devcontainer struct {
	Spec struct {
		Name              string
		RemoteUser        string `mapstructure:"remoteUser"`
		ContainerUser     string `mapstructure:"containerUser"`
		DockerComposeFile string `mapstructure:"dockerComposeFile"`
		Service           string `mapstructure:"service"`
		WorkspaceFolder   string `mapstructure:"workspaceFolder"`
	}
	FilePath string
}

type Config struct {
	Neovim struct {
		ConfigURI string `mapstructure:"config_uri"`
	}
	DevcontainerFile string `mapstructure:"devcontainer_file"`

	FilePath string `mapstructure:"-"`
}

func (config *Config) GetDevcontainerFilePath() string {
	if config.DevcontainerFile[:2] == "./" {
		return filepath.Join(filepath.Dir(config.FilePath), config.DevcontainerFile)
	}
	return config.DevcontainerFile
}

func (config *Config) GetConfigURI() (*url.URL, error) {
	return url.Parse(config.Neovim.ConfigURI)
}

type ConfigViper struct {
	Config Config
	Viper  *viper.Viper
}

const ConfigFileBaseName = "nvim-mindevc"
const DefaultConfigFile = "." + ConfigFileBaseName + ".yaml"

func LoadConfig(loadConfigFile string) (ConfigViper, error) {
	var configConfig Config
	var configViperViper = viper.New()

	configViperViper.SetDefault("neovim.config_uri", "file://~/.config/nvim")

	if loadConfigFile != "" {
		configViperViper.SetConfigFile(loadConfigFile)
		if err := configViperViper.ReadInConfig(); err != nil {
			return ConfigViper{}, err
		}
		configConfig.FilePath = loadConfigFile
	} else {
		var configFile = filepath.Join(".", DefaultConfigFile)
		configViperViper.SetConfigFile(configFile)
		if err := configViperViper.ReadInConfig(); err == nil {
			configConfig.FilePath = configFile
		} else {
			slog.Debug("Could not read config file", "path", configFile, "error", err)

			configViperViper.SetConfigName(ConfigFileBaseName)
			configViperViper.SetConfigType("yaml")
			configViperViper.AddConfigPath(".devcontainer")

			if home, err := os.UserHomeDir(); err == nil {
				configViperViper.AddConfigPath(filepath.Join(home, ".config"))
			}

			if err := configViperViper.ReadInConfig(); err != nil {
				slog.Debug("Could not read config files", "error", err)
			}

			configConfig.FilePath = configViperViper.ConfigFileUsed()
		}
	}

	configViperViper.SetEnvPrefix("NVIM_MINDEVC")
	configViperViper.AutomaticEnv()

	configViperViper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := configViperViper.Unmarshal(&configConfig); err != nil {
		return ConfigViper{}, err
	}

	return ConfigViper{
		Config: configConfig,
		Viper:  configViperViper,
	}, nil
}

func LoadDevcontainer(loadDevcontainerFile string) (Devcontainer, error) {
	var devcontainer Devcontainer

	var devcontainerViper = viper.New()
	if loadDevcontainerFile != "" {
		devcontainerViper.SetConfigFile(loadDevcontainerFile)
		if err := devcontainerViper.ReadInConfig(); err != nil {
			return Devcontainer{}, err
		}
		devcontainer.FilePath = loadDevcontainerFile
	} else {
		var devcontainerExists = false
		var devcontainerFiles = []string{
			filepath.Join(".devcontainer", "devcontainer.json"),
			".devcontainer.json"}

		for _, filePath := range devcontainerFiles {
			devcontainerViper.SetConfigFile(filePath)
			if err := devcontainerViper.ReadInConfig(); err == nil {
				devcontainerExists = true
				devcontainer.FilePath = filePath
				break
			}
		}

		if !devcontainerExists {
			return Devcontainer{}, fmt.Errorf("no devcontainer file found")
		}
	}

	if err := devcontainerViper.Unmarshal(&devcontainer.Spec); err != nil {
		return Devcontainer{}, err
	}

	return devcontainer, nil
}
