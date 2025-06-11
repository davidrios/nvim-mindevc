package setup

import (
	"compress/gzip"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func DownloadAndExtractNeovim(workDir string, tag string, noCache bool) (string, error) {
	slog.Debug("downloading neovim", "tag", tag)

	neovimSourceFile := filepath.Join(workDir, fmt.Sprintf("neovim-%s.tar.gz", tag))
	if _, err := os.Stat(neovimSourceFile); err != nil || noCache {
		tmpFile := neovimSourceFile + ".tmp"
		err := DownloadFileHttp(fmt.Sprintf("https://github.com/neovim/neovim/archive/refs/tags/%s.tar.gz", tag), tmpFile)
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

	err = extractTar(fileReader, workDir)
	if err != nil {
		return "", fmt.Errorf("failed to extract tar: %w", err)
	}

	slog.Debug("downloaded and extracted")
	return filepath.Join(workDir, "neovim-"+tag), nil
}

func CompileAndPackNeovim(zigBin string, neovimSrc string, arch string, dataDir string) (string, error) {
	slog.Debug("compiling neovim for", "arch", arch)

	if err := ReplaceInFile(filepath.Join(neovimSrc, "src/nvim/os/stdpaths.c"), "/usr/local/share/:/usr/share/", dataDir); err != nil {
		return "", err
	}

	cmd := exec.Command(zigBin, "build", "nvim", fmt.Sprintf("-Dtarget=%s-linux-musl", arch))
	cmd.Dir = neovimSrc
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error compiling, %w, %s", err, cmd.Stderr)
	}

	return "", fmt.Errorf("error")
}

func TestWithRedis(t *testing.T) {
}
