package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDownloadAndCompileNeovim(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping neovim compile test")
	}

	tempDir := filepath.Join(os.TempDir(), "nvim-mindev-tests", "neovim")
	err := os.MkdirAll(tempDir, 0o700)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %s", err)
	}

	neovimSrc, err := DownloadAndExtractNeovim(tempDir, "nightly", false)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err = DownloadAndExtractLocalTools(tempDir); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	zigBin := filepath.Join(tempDir, "bin", "zig")

	cmd := exec.Command("uname", "-m")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("could not get current OS: %s", err)
	}

	arch := strings.TrimSpace(string(output))

	archive, err := CompileAndPackNeovim(zigBin, neovimSrc, arch, filepath.Join(tempDir, "compiled")+"/")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	fmt.Print("done ", archive)
}
