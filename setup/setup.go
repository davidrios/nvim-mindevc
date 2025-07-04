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

func Setup(myConfig config.ConfigViper, devcontainer config.Devcontainer, skipSelfBinary bool) error {
	if devcontainer.Spec.DockerComposeFile == "" {
		return fmt.Errorf("dockerComposeFile property from devcontainer file must not be empty")
	}

	if devcontainer.Spec.Service == "" {
		return fmt.Errorf("service property from devcontainer file must not be empty")
	}

	if devcontainer.Spec.RemoteUser == "" {
		return fmt.Errorf("remoteUser property from devcontainer file must not be empty")
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
		err = composeFile.CpToService(serviceName, downloadedFile, filepath.Join(uploadDir, filepath.Base(downloadedFile)), docker.CpToServiceOptions{})
		if err != nil {
			return err
		}
		if uncFile, _ := UncompressTool(myConfig.Config.Tools[toolName].Archives[arch].Type, downloadedFile); uncFile != "" {
			_ = composeFile.CpToService(serviceName, uncFile, filepath.Join(uploadDir, filepath.Base(uncFile)), docker.CpToServiceOptions{})
		}
		slog.Debug("copied tool to remote", "file", downloadedFile)
	}

	remoteBinary := filepath.Join(uploadDir, "nvim-mindevc")

	if !skipSelfBinary {
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
			err = composeFile.CpToService(serviceName, myPath, remoteBinary, docker.CpToServiceOptions{})
			if err != nil {
				return err
			}
			slog.Debug("copied self binary", "p", myPath)
		} else {
			slog.Warn("cannot use self binary, incompatible remote os and/or architecture")
		}
	}

	_, err = composeFile.Exec(serviceName, docker.ExecParams{
		Args: []string{"chmod", "+x", remoteBinary},
		User: "root",
	})
	if err != nil {
		return err
	}

	myConfig.Viper.Set("remote.user", devcontainer.Spec.RemoteUser)

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
	err = composeFile.CpToService(serviceName, file.Name(), remoteConfig, docker.CpToServiceOptions{})
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
	err = composeFile.CpToService(serviceName, file.Name(), filepath.Join(myConfig.Config.Remote.Workdir, "cacert.pem"), docker.CpToServiceOptions{})
	if err != nil {
		return err
	}

	output, err := composeFile.Exec(serviceName, docker.ExecParams{
		Args: []string{"sh", "-l", "-c", "echo $HOME"},
		User: devcontainer.Spec.RemoteUser,
	})
	if err != nil {
		return fmt.Errorf("error getting remote user home: %w", err)
	}
	remoteHome := strings.TrimSpace(output)
	if remoteHome == "/" {
		return fmt.Errorf("error getting remote user home, got '/'")
	}

	configUri, err := myConfig.Config.GetConfigURI()
	if err != nil {
		return err
	}
	slog.Debug("got config uri", "uri", configUri, "scheme", configUri.Scheme)

	switch configUri.Scheme {
	case "file":
		configPath, err := config.ExpandHome(myConfig.Config.Neovim.ConfigURI[len("file://"):])
		if err != nil {
			return err
		}
		slog.Debug("nvim config path", "p", configPath)

		output, err = composeFile.Exec(serviceName, docker.ExecParams{
			Args: []string{"sh", "-c",
				fmt.Sprintf(
					"mkdir -p '%s/.config' && chown -R '%s' '%s' && test -d '%s/.config/nvim' || echo -n 'nvim_not_found'",
					remoteHome,
					devcontainer.Spec.RemoteUser,
					remoteHome,
					remoteHome)},
			User: "root",
		})
		if err != nil {
			return fmt.Errorf("error configuring user home: %w", err)
		}

		if output == "nvim_not_found" {
			err = composeFile.CpToService(
				serviceName, configPath, filepath.Join(remoteHome, ".config", "nvim"),
				docker.CpToServiceOptions{FollowLink: true})
			if err != nil {
				return err
			}
		} else {
			slog.Warn("remote nvim config dir exists, not overwritting...")
		}
	default:
		return fmt.Errorf("invalid nvim config uri, skipping")
	}

	slog.Info("running remote setup, this might take a while...")
	output, err = composeFile.Exec(serviceName, docker.ExecParams{
		Args: []string{remoteBinary, "-v", "-c", remoteConfig, "remote-setup"},
		User: "root",
	})
	if err != nil {
		return err
	}
	slog.Debug("out", "o", output)

	slog.Info("all done")

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

	gitLink := filepath.Join(myConfig.Config.Remote.Workdir, "bin", "git")
	if _, err := os.Lstat(gitLink); err == nil {
		if err := os.Remove(gitLink); err != nil {
			return fmt.Errorf("failed to remove existing symlink %s: %w", gitLink, err)
		}
	}
	gitLinkTarget := filepath.Join(myConfig.Config.Remote.Workdir, "tools", "_download", "nvim-mindevc")
	if err := os.Symlink(gitLinkTarget, gitLink); err != nil {
		return fmt.Errorf("failed to create symlink %s -> %s: %w", gitLink, gitLinkTarget, err)
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

	zigBin := filepath.Join(toolsDir, "zig", config.ZigTool.Archives[arch].Links[config.DefaultZigLink])

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

	if myConfig.Config.Remote.ExtraBashRc != "" {
		output, err := exec.Command("getent", "passwd", myConfig.Config.Remote.User).Output()
		if err != nil {
			return err
		}
		userHome := strings.Split(string(output), ":")[5]
		if userHome == "/" || userHome == "" {
			return fmt.Errorf("error getting remote user home")
		}
		extraRc := filepath.Join(userHome, ".bashrc_extra")

		file, err := os.Create(extraRc)
		if err != nil {
			return fmt.Errorf("error opening file: %w", err)
		}

		if _, err := io.Copy(file, bytes.NewReader([]byte(myConfig.Config.Remote.ExtraBashRc))); err != nil {
			file.Close()
			return fmt.Errorf("unexpected error: %s", err)
		}
		file.Close()

		rcFile := filepath.Join(userHome, ".bashrc")

		file, err = os.OpenFile(rcFile, os.O_CREATE|os.O_RDWR, 0o600)
		if err != nil {
			return err
		}
		defer file.Close()

		lineToAdd := ". " + extraRc
		hasLine, err := utils.FileContainsLine(file, lineToAdd)
		if err != nil {
			return err
		}
		if !hasLine {
			_, err = file.Seek(0, 2)
			if err != nil {
				return err
			}
			_, err = file.WriteString("\n" + lineToAdd + "\n")
			if err != nil {
				return err
			}
		}
		file.Close()

		err = exec.Command("chown", "-R", myConfig.Config.Remote.User, userHome).Run()
		if err != nil {
			return err
		}
	}

	return nil
}
