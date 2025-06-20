package setup

import (
	"compress/gzip"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/davidrios/nvim-mindevc/utils"
)

func IsAlpine() (bool, error) {
	cmd := exec.Command("uname", "-m")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			slog.Debug("cmd error", "stderr", exitErr.Stderr)
		}
		return false, fmt.Errorf("error executing: %w", err)
	}
	_arch := strings.TrimSpace(string(output))
	_, err = os.Stat(fmt.Sprintf("/lib/ld-musl-%s.so.1", _arch))
	return err == nil, nil
}

func DownloadAndExtractNeovim(workDir string, tag string, noCache bool) (string, error) {
	slog.Debug("downloading neovim", "tag", tag)

	neovimSourceFile := filepath.Join(workDir, fmt.Sprintf("neovim-%s.tar.gz", tag))
	if _, err := os.Stat(neovimSourceFile); err != nil || noCache {
		tmpFile := neovimSourceFile + ".tmp"
		err := utils.DownloadFileHttp(fmt.Sprintf("https://github.com/neovim/neovim/archive/refs/tags/%s.tar.gz", tag), tmpFile)
		if err != nil {
			return "", err
		}
		err = os.Rename(tmpFile, neovimSourceFile)
		if err != nil {
			return "", err
		}
	}

	toolFile, err := os.Open(neovimSourceFile)
	if err != nil {
		return "", fmt.Errorf("could not open downloaded file: %w", err)
	}
	defer toolFile.Close()

	fileReader, err := gzip.NewReader(toolFile)
	if err != nil {
		return "", err
	}

	err = utils.ExtractTar(fileReader, workDir)
	if err != nil {
		return "", fmt.Errorf("failed to extract tar: %w", err)
	}

	slog.Debug("downloaded and extracted")
	neovimSrc := filepath.Join(workDir, "neovim-"+tag)

	isAlpine, err := IsAlpine()
	if err != nil {
		return "", err
	}
	if isAlpine {
		main_c := filepath.Join(neovimSrc, "src/nvim/main.c")

		if err = utils.ReplaceInFile(
			main_c,
			"main(int argc, char **argv)",
			"main(int argc, char **argv, char **envp)",
		); err != nil {
			return "", err
		}
		if err = utils.ReplaceInFile(
			main_c,
			"argv0 = argv[0];",
			"extern char **__environ; __environ = envp; argv0 = argv[0];",
		); err != nil {
			return "", err
		}

	}

	return neovimSrc, nil
}

func CompileNeovim(zigBin string, neovimSrc string) error {
	nvimBin := filepath.Join(neovimSrc, "zig-out", "bin", "nvim")

	cmd := exec.Command(nvimBin, "--clean", "-es", "-c", "call writefile(['hello'], '.imalive')")
	cmd.Dir = neovimSrc
	cmd.Env = append(cmd.Env, "VIM="+neovimSrc)
	if err := cmd.Run(); err != nil {
		slog.Info("compiling neovim, this may take a while...")

		cmd = exec.Command(zigBin, "build", "nvim", "--release=fast", "-Dluajit")
		cmd.Dir = neovimSrc
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error compiling, %w, %s", err, cmd.Stderr)
		}
		slog.Info("done")
	}

	return nil
}

func TarNeovim(neovimSrc string, destFile string) error {
	return nil
}
