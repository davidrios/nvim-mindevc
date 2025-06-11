package config

import (
	"fmt"
	"log/slog"
	"maps"
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
	ArchiveTypeBinGz  ConfigToolArchiveType = "bin.gz"
	ArchiveTypeBinBz2 ConfigToolArchiveType = "bin.bz2"
	ArchiveTypeBinXz  ConfigToolArchiveType = "bin.xz"
)

var ValidArchiveTypes map[string]struct{} = map[string]struct{}{
	string(ArchiveTypeBin):    {},
	string(ArchiveTypeBinGz):  {},
	string(ArchiveTypeBinBz2): {},
	string(ArchiveTypeBinXz):  {},
	string(ArchiveTypeZip):    {},
	string(ArchiveTypeTarGz):  {},
	string(ArchiveTypeTarBz2): {},
	string(ArchiveTypeTarXz):  {},
}

func (archiveType *ConfigToolArchiveType) IsValid() bool {
	_, ok := ValidArchiveTypes[string(*archiveType)]
	return ok
}

func (archiveType *ConfigToolArchiveType) IsTar() bool {
	archiveTypeVal := *archiveType
	return archiveTypeVal == ArchiveTypeTarGz || archiveTypeVal == ArchiveTypeTarBz2 || archiveTypeVal == ArchiveTypeTarXz
}

func (archiveType *ConfigToolArchiveType) IsGBXZCompressed() bool {
	archiveTypeVal := *archiveType
	return archiveTypeVal == ArchiveTypeTarGz || archiveTypeVal == ArchiveTypeTarBz2 ||
		archiveTypeVal == ArchiveTypeTarXz || archiveTypeVal == ArchiveTypeBinGz ||
		archiveTypeVal == ArchiveTypeBinBz2 || archiveTypeVal == ArchiveTypeBinXz
}

type ConfigToolArchive struct {
	Url   string
	Hash  string
	Type  ConfigToolArchiveType
	Links map[string]string
}

type ConfigTool struct {
	Source   ConfigToolSource
	Archives map[ConfigToolArch]ConfigToolArchive
}

type ConfigTools map[string]ConfigTool

type Config struct {
	Neovim struct {
		ConfigURI string `mapstructure:"config_uri"`
	}
	InstallTools     []string `mapstructure:"install_tools"`
	DevcontainerFile string   `mapstructure:"devcontainer_file"`
	Tools            ConfigTools
	CacheDir         string `mapstructure:"cache_dir"`
	Remote           struct {
		Workdir string
	}

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
					Url:  "https://github.com/sharkdp/fd/releases/download/v10.2.0/fd-v10.2.0-x86_64-unknown-linux-musl.tar.gz",
					Hash: "d9bfa25ec28624545c222992e1b00673b7c9ca5eb15393c40369f10b28f9c932",
					Type: ArchiveTypeTarGz,
					Links: map[string]string{
						"/usr/local/bin/fd": "fd-v10.2.0-x86_64-unknown-linux-musl/fd",
					},
				},
				ToolArch_aarch64: {
					Url:  "https://github.com/sharkdp/fd/releases/download/v10.2.0/fd-v10.2.0-aarch64-unknown-linux-musl.tar.gz",
					Hash: "4e8e596646d047d904f2c5ca74b39dccc69978b6e1fb101094e534b0b59c1bb0",
					Type: ArchiveTypeTarGz,
					Links: map[string]string{
						"/usr/local/bin/fd": "fd-v10.2.0-aarch64-unknown-linux-musl/fd",
					},
				},
			},
		},
		"ripgrep": {
			Source: ToolSourceArchive,
			Archives: map[ConfigToolArch]ConfigToolArchive{
				ToolArch_x86_64: {
					Url:  "https://github.com/BurntSushi/ripgrep/releases/download/14.1.1/ripgrep-14.1.1-x86_64-unknown-linux-musl.tar.gz",
					Hash: "4cf9f2741e6c465ffdb7c26f38056a59e2a2544b51f7cc128ef28337eeae4d8e",
					Type: ArchiveTypeTarGz,
					Links: map[string]string{
						"/usr/local/bin/rg": "ripgrep-14.1.1-x86_64-unknown-linux-musl/rg",
					},
				},
				ToolArch_aarch64: {
					Url:  "https://github.com/BurntSushi/ripgrep/releases/download/14.1.1/ripgrep-14.1.1-armv7-unknown-linux-musleabi.tar.gz",
					Hash: "e6512cb9d3d53050022b9236edd2eff4244cea343a451bfb3c008af23d0000e5",
					Type: ArchiveTypeTarGz,
					Links: map[string]string{
						"/usr/local/bin/rg": "ripgrep-14.1.1-armv7-unknown-linux-musl/rg",
					},
				},
			},
		},
		"gosu": {
			Source: ToolSourceArchive,
			Archives: map[ConfigToolArch]ConfigToolArchive{
				ToolArch_x86_64: {
					Url:  "https://github.com/tianon/gosu/releases/download/1.17/gosu-amd64",
					Hash: "bbc4136d03ab138b1ad66fa4fc051bafc6cc7ffae632b069a53657279a450de3",
					Type: ArchiveTypeBin,
					Links: map[string]string{
						"/usr/local/bin/gosu": "$bin",
					},
				},
				ToolArch_aarch64: {
					Url:  "https://github.com/tianon/gosu/releases/download/1.17/gosu-arm64",
					Hash: "c3805a85d17f4454c23d7059bcb97e1ec1af272b90126e79ed002342de08389b",
					Type: ArchiveTypeBin,
					Links: map[string]string{
						"/usr/local/bin/gosu": "$bin",
					},
				},
			},
		},
		"curl": {
			Source: ToolSourceArchive,
			Archives: map[ConfigToolArch]ConfigToolArchive{
				ToolArch_x86_64: {
					Url:  "https://github.com/stunnel/static-curl/releases/download/8.14.1/curl-linux-x86_64-musl-8.14.1.tar.xz",
					Hash: "0b4622d9df4fd282b5a2d222e4e0146fc409053ee15ee1979784f6c8a56cf573",
					Type: ArchiveTypeTarXz,
					Links: map[string]string{
						"/usr/local/bin/curl":  "curl",
						"/usr/local/bin/trurl": "trurl",
					},
				},
				ToolArch_aarch64: {
					Url:  "https://github.com/stunnel/static-curl/releases/download/8.14.1/curl-linux-aarch64-musl-8.14.1.tar.xz",
					Hash: "e0fecb5ecaba101b4b560f1035835770e7d1c151416ee84e18c813ba32b9d1dd",
					Type: ArchiveTypeTarXz,
					Links: map[string]string{
						"/usr/local/bin/curl":  "curl",
						"/usr/local/bin/trurl": "trurl",
					},
				},
			},
		},
		"zig": {
			Source: ToolSourceArchive,
			Archives: map[ConfigToolArch]ConfigToolArchive{
				ToolArch_x86_64: {
					Url:  "https://ziglang.org/download/0.14.1/zig-x86_64-linux-0.14.1.tar.xz",
					Hash: "24aeeec8af16c381934a6cd7d95c807a8cb2cf7df9fa40d359aa884195c4716c",
					Type: ArchiveTypeTarXz,
					Links: map[string]string{
						"/usr/local/bin/zig": "zig-x86_64-linux-0.14.1/zig",
					},
				},
				ToolArch_aarch64: {
					Url:  "https://ziglang.org/download/0.14.1/zig-aarch64-linux-0.14.1.tar.xz",
					Hash: "f7a654acc967864f7a050ddacfaa778c7504a0eca8d2b678839c21eea47c992b",
					Type: ArchiveTypeTarXz,
					Links: map[string]string{
						"/usr/local/bin/zig": "zig-aarch64-linux-0.14.1/zig",
					},
				},
			},
		},
	})
	configViperViper.SetDefault("install_tools", []string{"fd", "ripgrep", "gosu", "curl", "zig"})
	configViperViper.SetDefault("neovim.config_uri", "file://~/.config/nvim")
	configViperViper.SetDefault("remote.workdir", "/opt/nvim-mindevc")
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

type NvimMindevcTool struct {
	InstallTools []string
	Tools        ConfigTools
}

const NvimMindevcVersion string = "nightly"

func WithNvimMindevcTool(config Config) NvimMindevcTool {
	installTools := config.InstallTools
	tools := make(ConfigTools, len(config.Tools)+1)
	maps.Copy(tools, config.Tools)

	installTools = append(installTools, "nvim-mindevc")
	tools["nvim-mindevc"] = ConfigTool{
		Source: ToolSourceArchive,
		Archives: map[ConfigToolArch]ConfigToolArchive{
			ToolArch_aarch64: {
				Url:   fmt.Sprintf("https://github.com/davidrios/nvim-mindevc/releases/download/%s/nvim-mindevc-linux-aarch64.gz", NvimMindevcVersion),
				Hash:  "e33597db93b4d6146c953c8d72f0adcbc62f8b8acdeb68e8a6512369331f84d7",
				Type:  ArchiveTypeBinGz,
				Links: map[string]string{"/usr/local/bin/nvim-mindevc": "$bin"},
			},
			ToolArch_x86_64: {
				Url:   fmt.Sprintf("https://github.com/davidrios/nvim-mindevc/releases/download/%s/nvim-mindevc-linux-x86_64.gz", NvimMindevcVersion),
				Hash:  "3dc66a36fd12a1b0312fe78a34d968a2f92008982d624ea5e2765d40c0923722",
				Type:  ArchiveTypeBinGz,
				Links: map[string]string{"/usr/local/bin/nvim-mindevc": "$bin"},
			},
		},
	}

	return NvimMindevcTool{
		InstallTools: installTools,
		Tools:        tools,
	}
}
