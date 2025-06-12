package setup

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/davidrios/nvim-mindevc/config"
	"github.com/davidrios/nvim-mindevc/docker"
	"github.com/davidrios/nvim-mindevc/utils"
)

func Setup(myConfig config.ConfigViper, devcontainer config.Devcontainer, useSelfBinary bool) error {
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

	_arch, err := composeFile.Exec(serviceName, docker.ExecParams{
		Args: []string{"uname", "-m"},
		User: "root",
	})
	if err != nil {
		return err
	}

	_arch = strings.TrimSpace(_arch)
	slog.Debug("container arch", "v", _arch)
	arch := config.ConfigToolArch(_arch)

	withNvimMindevcTools := config.WithNvimMindevcTool(myConfig.Config)

	cacheDir, err := config.ExpandHome(myConfig.Config.CacheDir)
	if err != nil {
		return err
	}

	downloaded, err := DownloadTools(cacheDir,
		arch,
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

	for toolName, downloadedFile := range downloaded {
		err = composeFile.CpToService(serviceName, downloadedFile, filepath.Join(uploadDir, filepath.Base(downloadedFile)))
		if err != nil {
			return err
		}
		if uncFile, _ := UncompressTool(myConfig.Config.Tools[toolName].Archives[arch].Type, downloadedFile); uncFile != "" {
			_ = composeFile.CpToService(serviceName, uncFile, filepath.Join(uploadDir, filepath.Base(uncFile)))
		}
		slog.Debug("copied tool to remote", "file", downloadedFile)
	}

	remoteBinary := filepath.Join(uploadDir, "nvim-mindevc")

	if useSelfBinary {
		cmd := exec.Command("uname", "-sm")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("could not get current OS: %w", err)
		}

		osArch := strings.TrimSpace(string(output))
		if osArch == fmt.Sprintf("Linux %s", _arch) {
			myPath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("could not get current binary: %w", err)
			}
			err = composeFile.CpToService(serviceName, myPath, remoteBinary)
			if err != nil {
				return err
			}
			slog.Debug("copied self binary", "p", myPath)
		} else {
			slog.Warn("cannot use self binary, incompatible remote os and/or architecture")
		}
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
	remoteConfig := filepath.Join(myConfig.Config.Remote.Workdir, "config.yaml")
	err = composeFile.CpToService(serviceName, file.Name(), remoteConfig)
	if err != nil {
		return err
	}

	file, err = os.CreateTemp("", "")
	if err != nil {
		return fmt.Errorf("error opening temp file: %w", err)
	}
	defer os.Remove(file.Name())
	file.Close()

	err = utils.DownloadFileHttp("https://curl.se/ca/cacert.pem", file.Name())
	if err != nil {
		return fmt.Errorf("error downloading cacerts: %w", err)
	}
	err = composeFile.CpToService(serviceName, file.Name(), filepath.Join(myConfig.Config.Remote.Workdir, "cacert.pem"))
	if err != nil {
		return err
	}

	url, err := myConfig.Config.GetConfigURI()
	if err != nil {
		return err
	}

	slog.Debug("got URL", "url", url, "scheme", url.Scheme)

	slog.Info("running remote setup, this might take a while...")
	output, err := composeFile.Exec(serviceName, docker.ExecParams{
		Args: []string{remoteBinary, "-v", "-c", remoteConfig, "remote-setup"},
		User: "root",
	})
	if err != nil {
		return err
	}
	slog.Debug("out", "o", output)

	return nil
}

func RemoteSetup(myConfig config.ConfigViper) error {
	cmd := exec.Command("uname", "-m")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			slog.Debug("cmd error", "stderr", exitErr.Stderr)
		}
		return fmt.Errorf("error executing: %w", err)
	}

	_arch := strings.TrimSpace(string(output))
	arch := config.ConfigToolArch(_arch)
	slog.Debug("container arch", "v", arch)

	downloaded, err := DownloadTools(myConfig.Config.Remote.Workdir,
		arch,
		myConfig.Config.InstallTools,
		myConfig.Config.Tools,
	)
	if err != nil {
		return err
	}

	extracted, err := ExtractTools(
		arch,
		myConfig.Config.InstallTools,
		myConfig.Config.Tools,
		downloaded,
	)
	if err != nil {
		return err
	}

	err = LinkTools(
		arch,
		myConfig.Config.InstallTools,
		myConfig.Config.Tools,
		extracted,
	)
	if err != nil {
		return err
	}

	caFile := "/etc/ssl/certs/ca-certificates.crt"
	if err := os.MkdirAll(filepath.Dir(caFile), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(caFile); err != nil {
		fp, err := os.Create(caFile)
		if err != nil {
			return err
		}
		defer fp.Close()

		sfp, err := os.Open(filepath.Join(myConfig.Config.Remote.Workdir, "cacert.pem"))
		if err != nil {
			return err
		}
		defer sfp.Close()

		if _, err := io.Copy(fp, sfp); err != nil {
			return err
		}
	}

	toolsDir := filepath.Join(myConfig.Config.Remote.Workdir, "tools", _arch)

	neovimDir := filepath.Join(myConfig.Config.Remote.Workdir, "neovim")
	if err := os.MkdirAll(neovimDir, 0o755); err != nil {
		return err
	}

	neovimSrc, err := DownloadAndExtractNeovim(neovimDir, "nightly", false)
	if err != nil {
		return err
	}

	zigBin := filepath.Join(toolsDir, "zig", config.ZigTool.Archives[arch].Links["/usr/local/bin/zig"])

	err = CompileNeovim(zigBin, neovimSrc)
	if err != nil {
		return err
	}

	nvimRun := fmt.Sprintf(`#!/bin/sh
VIM="%s" "%s" "$@"`, neovimSrc, filepath.Join(neovimSrc, "zig-out", "bin", "nvim"))
	fp, err := os.Create(myConfig.Config.Neovim.Runscript)
	if err != nil {
		return err
	}
	defer fp.Close()
	_, err = fp.WriteString(nvimRun)
	if err != nil {
		return err
	}
	if err := os.Chmod(myConfig.Config.Neovim.Runscript, 0o755); err != nil {
		return err
	}

	return nil
}
