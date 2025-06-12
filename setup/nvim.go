package setup

import (
	"compress/gzip"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/davidrios/nvim-mindevc/utils"
)

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
	return filepath.Join(workDir, "neovim-"+tag), nil
}

func CompileNeovim(zigBin string, neovimSrc string) error {
	nvimBin := filepath.Join(neovimSrc, "zig-out", "bin", "nvim")

	cmd := exec.Command(nvimBin, "--clean", "-es", "-c", "call writefile(['hello'], '.imalive')")
	cmd.Dir = neovimSrc
	cmd.Env = append(cmd.Env, "VIM="+neovimSrc)
	if err := cmd.Run(); err != nil {
		slog.Info("compiling neovim, this may take a while...")
		cmd = exec.Command(zigBin, "build", "nvim", "--release=fast")
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

func TestWithRedis(t *testing.T) {
}
