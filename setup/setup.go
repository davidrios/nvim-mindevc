package setup

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/davidrios/nvim-mindevc/config"
	"github.com/davidrios/nvim-mindevc/docker"
	"gopkg.in/yaml.v3"
)

func Setup(myConfig config.ConfigViper, devcontainer config.Devcontainer) error {
	if devcontainer.Spec.DockerComposeFile == "" {
		return fmt.Errorf("dockerComposeFile property from devcontainer file must not be empty")
	}

	if devcontainer.Spec.Service == "" {
		return fmt.Errorf("service property from devcontainer file must not be empty")
	}

	composeFile, err := docker.LoadComposeFile(devcontainer)
	if err != nil {
		return fmt.Errorf("error loading compose file: %w", err)
	}
	slog.Debug("composeFile", "v", composeFile)

	serviceName := devcontainer.Spec.Service
	if _, ok := composeFile.Spec.Services[serviceName]; !ok {
		return fmt.Errorf("compose file does not contain service '%s'", serviceName)
	}

	arch, err := composeFile.Exec(serviceName, docker.ExecParams{
		Args: []string{"uname", "-m"},
		User: "root",
	})
	if err != nil {
		return err
	}

	arch = strings.TrimSpace(arch)
	slog.Debug("container arch", "v", arch)

	withNvimMindevcTools := config.WithNvimMindevcTool(myConfig.Config)

	downloaded, err := DownloadTools(myConfig.Config.CacheDir,
		config.ConfigToolArch(arch),
		withNvimMindevcTools.InstallTools,
		withNvimMindevcTools.Tools,
	)
	if err != nil {
		return err
	}

	uploadDir := filepath.Join(myConfig.Config.Remote.Workdir, "tools", "_download")
	_, err = composeFile.Exec(serviceName, docker.ExecParams{
		Args: []string{"mkdir", "-p", uploadDir},
		User: "root",
	})
	if err != nil {
		return err
	}

	for _, tool := range downloaded {
		err = composeFile.CpToService(serviceName, tool, filepath.Join(uploadDir, filepath.Base(tool)))
		if err != nil {
			return err
		}
		slog.Debug("copied tool to remote", "file", tool)
	}

	yamlData, err := yaml.Marshal(myConfig.Viper.AllSettings())
	if err != nil {
		return fmt.Errorf("Failed to marshal config to YAML: %w", err)
	}

	file, err := os.CreateTemp("", "")
	if err != nil {
		return fmt.Errorf("error opening temp file: %w", err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	if _, err := io.Copy(file, bytes.NewReader(yamlData)); err != nil {
		return fmt.Errorf("unexpected error: %s", err)
	}
	err = composeFile.CpToService(serviceName, file.Name(), filepath.Join(myConfig.Config.Remote.Workdir, "config.yaml"))
	if err != nil {
		return err
	}

	url, err := myConfig.Config.GetConfigURI()
	if err != nil {
		return err
	}

	slog.Debug("got URL", "url", url, "scheme", url.Scheme)

	return nil
}
