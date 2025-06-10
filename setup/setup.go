package setup

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/davidrios/nvim-mindevc/config"
	"github.com/davidrios/nvim-mindevc/docker"
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

	if _, err := DownloadTools(myConfig.Config.CacheDir,
		config.ConfigToolArch(arch),
		withNvimMindevcTools.InstallTools,
		withNvimMindevcTools.Tools,
	); err != nil {
		return err
	}

	url, err := myConfig.Config.GetConfigURI()
	if err != nil {
		return err
	}

	slog.Debug("got URL", "url", url, "scheme", url.Scheme)

	return nil
}
