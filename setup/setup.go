package setup

import (
	"fmt"
	"log/slog"

	"github.com/davidrios/nvim-mindevc/config"
	"github.com/davidrios/nvim-mindevc/docker"
)

func Setup(config config.ConfigViper, devcontainer config.Devcontainer) error {
	url, err := config.Config.GetConfigURI()
	if err != nil {
		return err
	}

	if devcontainer.Spec.DockerComposeFile == "" {
		return fmt.Errorf("dockerComposeFile property from devcontainer file must not be empty")
	}

	if devcontainer.Spec.Service == "" {
		return fmt.Errorf("service property from devcontainer file must not be empty")
	}

	var composeFile docker.ComposeFile
	composeFile, err = docker.LoadComposeFile(devcontainer)
	if err != nil {
		return fmt.Errorf("error loading compose file: %w", err)
	}
	slog.Debug("composeFile", "v", composeFile)

	serviceName := devcontainer.Spec.Service
	if _, ok := composeFile.Spec.Services[serviceName]; !ok {
		return fmt.Errorf("compose file does not contain service '%s'", serviceName)
	}

	if output, err := composeFile.Exec(serviceName, docker.ExecParams{
		Args: []string{
			"bash", "-c",
			`echo aaa
echo bbb && cat <<EOF
hello
EOF`}}); err == nil {
		slog.Debug("exec", "o", output)
	}

	if output, err := composeFile.Exec(serviceName, docker.ExecParams{Args: []string{"id", "-u"}, User: "root:root"}); err == nil {
		slog.Debug("exec", "o", output)
	}

	// servicePs, err := composeFile.Ps(serviceName)
	// if err != nil {
	// 	return err
	// }
	//
	// slog.Debug("servicePs", "v", servicePs)
	//
	// if servicePs.State != "running" {
	// 	return fmt.Errorf("container is not running")
	// }

	slog.Debug("got URL", "url", url, "scheme", url.Scheme)
	return nil
}
