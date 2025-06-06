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

type ConfigToolSource string

const (
	ToolSourceArchive ConfigToolSource = "archive"
	ToolSourceGitRepo ConfigToolSource = "git-repo"
)

type ConfigToolArch string

const (
	ToolArch_x86_64  ConfigToolArch = "x86_64"
	ToolArch_aarch64 ConfigToolArch = "aarch64"
)

type ConfigToolArchiveType string

const (
	ArchiveTypeTarXz  ConfigToolArchiveType = "tar.xz"
	ArchiveTypeTarBz2 ConfigToolArchiveType = "tar.bz2"
	ArchiveTypeTarGz  ConfigToolArchiveType = "tar.gz"
	ArchiveTypeZip    ConfigToolArchiveType = "zip"
	ArchiveTypeBin    ConfigToolArchiveType = "bin"
)

type ConfigToolArchive struct {
	U string
	H string
	T ConfigToolArchiveType
}

type ConfigTool struct {
	Symlinks map[string]string
	Source   ConfigToolSource
	Archives map[ConfigToolArch]ConfigToolArchive
}

type ConfigTools map[string]ConfigTool

type Config struct {
	Neovim struct {
		ConfigURI string `mapstructure:"config_uri"`
	}
	InstallTools     []string `mapstructure:"install_tools"`
	UsrLocal         string   `mapstructure:"usr_local"`
	DevcontainerFile string   `mapstructure:"devcontainer_file"`
	Tools            ConfigTools
	CacheDir         string `mapstructure:"cache_dir"`

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

	configViperViper.SetDefault("tools", ConfigTools{
		"fd": {
			Source: ToolSourceArchive,
			Archives: map[ConfigToolArch]ConfigToolArchive{
				ToolArch_x86_64: {
					U: "https://github.com/sharkdp/fd/releases/download/v10.2.0/fd-v10.2.0-x86_64-unknown-linux-musl.tar.gz",
					H: "d9bfa25ec28624545c222992e1b00673b7c9ca5eb15393c40369f10b28f9c932",
					T: ArchiveTypeTarGz,
				},
				ToolArch_aarch64: {
					U: "https://github.com/sharkdp/fd/releases/download/v10.2.0/fd-v10.2.0-aarch64-unknown-linux-musl.tar.gz",
					H: "4e8e596646d047d904f2c5ca74b39dccc69978b6e1fb101094e534b0b59c1bb0",
					T: ArchiveTypeTarGz,
				},
			},
			Symlinks: map[string]string{
				"/usr/local/bin/fd": "fd",
			},
		},
		"ripgrep": {
			Source: ToolSourceArchive,
			Archives: map[ConfigToolArch]ConfigToolArchive{
				ToolArch_x86_64: {
					U: "https://github.com/BurntSushi/ripgrep/releases/download/14.1.1/ripgrep-14.1.1-x86_64-unknown-linux-musl.tar.gz",
					H: "4cf9f2741e6c465ffdb7c26f38056a59e2a2544b51f7cc128ef28337eeae4d8e",
					T: ArchiveTypeTarGz,
				},
				ToolArch_aarch64: {
					U: "https://github.com/BurntSushi/ripgrep/releases/download/14.1.1/ripgrep-14.1.1-armv7-unknown-linux-musleabi.tar.gz",
					H: "e6512cb9d3d53050022b9236edd2eff4244cea343a451bfb3c008af23d0000e5",
					T: ArchiveTypeTarGz,
				},
			},
			Symlinks: map[string]string{
				"/usr/local/bin/rg": "rg",
			},
		},
		"gosu": {
			Source: ToolSourceArchive,
			Archives: map[ConfigToolArch]ConfigToolArchive{
				ToolArch_x86_64: {
					U: "https://github.com/tianon/gosu/releases/download/1.17/gosu-amd64",
					H: "bbc4136d03ab138b1ad66fa4fc051bafc6cc7ffae632b069a53657279a450de3",
					T: ArchiveTypeBin,
				},
				ToolArch_aarch64: {
					U: "https://github.com/tianon/gosu/releases/download/1.17/gosu-arm64",
					H: "c3805a85d17f4454c23d7059bcb97e1ec1af272b90126e79ed002342de08389b",
					T: ArchiveTypeBin,
				},
			},
			Symlinks: map[string]string{
				"/usr/local/bin/gosu": "$bin",
			},
		},
	})
	configViperViper.SetDefault("install_tools", []string{"fd", "ripgrep", "gosu"})
	configViperViper.SetDefault("neovim.config_uri", "file://~/.config/nvim")
	configViperViper.SetDefault("usr_local", "/opt/nvim-mindevc")
	configViperViper.SetDefault("cache_dir", "~/.cache/nvim-mindevc")

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

func ExpandHome(pathstr string) (string, error) {
	if pathstr[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		pathstr = filepath.Join(home, pathstr[2:])
	}

	return pathstr, nil
}
