package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/davidrios/nvim-mindevc/config"
)

type ComposeFileService struct {
	User string
}

type ComposeFileSpec struct {
	Services map[string]ComposeFileService
}

type ComposeFile struct {
	Spec     ComposeFileSpec
	FilePath string
}

type ComposePsItem struct {
	ID      string
	Image   string
	Service string
	State   string
}

func (composeFile *ComposeFile) Ps(serviceName string) (ComposePsItem, error) {
	var servicePs ComposePsItem
	cmd := exec.Command("docker", "compose", "ps", "-a", "--format", "json")
	cmd.Dir = filepath.Dir(composeFile.FilePath)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			slog.Debug("cmd error", "stderr", exitErr.Stderr)
		}
		return servicePs, fmt.Errorf("error executing docker compose: %w", err)
	}
	slog.Debug("cmd output", "v", output[:200])

	if len(output) == 0 {
		return servicePs, fmt.Errorf("got empty compose output. try starting the compose project first")
	}

	lines := bytes.SplitSeq(output, []byte("\n"))
	for lineBytes := range lines {
		lineBytes = bytes.TrimSpace(lineBytes)

		var psItem ComposePsItem
		if err := json.Unmarshal(lineBytes, &psItem); err != nil {
			return servicePs, fmt.Errorf("error reading output: %w", err)
		}

		if psItem.Service == serviceName {
			return psItem, nil
		}
	}

	return servicePs, fmt.Errorf("service not found")
}

func (composeFile *ComposeFile) Exec(serviceName string, execParams ExecParams) (string, error) {
	cmdArgs := []string{"compose", "exec"}
	if execParams.Dettach {
		cmdArgs = append(cmdArgs, "--detach")
	}
	if len(execParams.Env) > 0 {
		for _, val := range execParams.Env {
			cmdArgs = append(cmdArgs, "--env", val)
		}
	}
	if execParams.Privileged {
		cmdArgs = append(cmdArgs, "--privileged")
	}
	if !execParams.Tty {
		cmdArgs = append(cmdArgs, "--no-TTY")
	}
	if execParams.User != "" {
		cmdArgs = append(cmdArgs, "--user", execParams.User)
	}
	if execParams.Workdir != "" {
		cmdArgs = append(cmdArgs, "--workdir", execParams.Workdir)
	}
	cmdArgs = append(cmdArgs, serviceName)
	cmdArgs = append(cmdArgs, execParams.Args...)
	cmd := exec.Command("docker", cmdArgs...)
	// cmd.Args = cmdArgs
	slog.Debug("cmdArgs", "v", cmd.Args)
	cmd.Dir = filepath.Dir(composeFile.FilePath)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			slog.Debug("cmd error", "stderr", exitErr.Stderr)
		}
		return "", fmt.Errorf("error executing docker compose: %w", err)
	}
	return string(output[:]), nil
}

func LoadComposeFile(devcontainer config.Devcontainer) (ComposeFile, error) {
	var composeFile ComposeFile

	readPath := filepath.Join(
		filepath.Dir(devcontainer.FilePath),
		devcontainer.Spec.DockerComposeFile)
	slog.Debug("read compose file from", "path", readPath)

	yamlFile, err := os.ReadFile(readPath)
	if err != nil {
		return composeFile, err
	}

	err = yaml.Unmarshal(yamlFile, &composeFile.Spec)
	if err != nil {
		return composeFile, err
	}

	composeFile.FilePath = readPath

	return composeFile, nil
}
