package setup

import (
	"os"
	"path/filepath"
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

	err = CompileNeovim(zigBin, neovimSrc)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}
