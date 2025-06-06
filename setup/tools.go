package setup

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/davidrios/nvim-mindevc/config"
)

func DownloadToolHttp(cacheDir string, rawUrl string, parsedUrl *url.URL, expectedHash string) (string, error) {
	cachedFilename := filepath.Join(cacheDir, expectedHash)

	if _, err := os.Stat(cachedFilename); err == nil {
		f, err := os.Open(cachedFilename)
		if err != nil {
			return "", err
		}
		defer f.Close()

		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return "", err
		}

		gotHash := fmt.Sprintf("%x", h.Sum(nil))
		if gotHash == expectedHash {
			slog.Debug("using cached file", "hash", expectedHash)
			return cachedFilename, nil
		}

		os.Remove(cachedFilename)
	}

	resp, err := http.Get(rawUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	tmpName := filepath.Join(cacheDir, expectedHash+".tmp")

	out, err := os.Create(tmpName)
	if err != nil {
		return "", err
	}
	defer out.Close()

	read, err := io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}
	if read == 0 {
		return "", fmt.Errorf("got empty file")
	}

	f, err := os.Open(tmpName)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	gotHash := fmt.Sprintf("%x", h.Sum(nil))
	if gotHash != expectedHash {
		return "", fmt.Errorf("hashes do not match")
	}

	err = os.Rename(tmpName, cachedFilename)
	if err != nil {
		return "", err
	}

	return cachedFilename, nil
}

func ExtractAndLinkTool(tool config.ConfigTool, fname string) error {
	return nil
}

func DownloadTools(myConfig config.Config, arch config.ConfigToolArch) error {
	slog.Debug("downloading tools")

	cacheDir, err := config.ExpandHome(myConfig.CacheDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(cacheDir, "tools"), 0o750); err != nil {
		return fmt.Errorf("error creating cache dir: %w", err)
	}

	for _, toolName := range myConfig.InstallTools {
		tool, ok := myConfig.Tools[toolName]
		if !ok {
			slog.Debug("tool does not exist", "tool", toolName)
			continue
		}

		switch tool.Source {
		case config.ToolSourceArchive:
			archive, ok := tool.Archives[arch]
			if !ok {
				slog.Warn("tool not found for arch", "tool", toolName, "arch", arch)
				continue
			}

			parsedUrl, err := url.Parse(archive.U)
			if err != nil {
				slog.Warn("invalid url for tool", "tool", toolName)
				continue
			}

			switch parsedUrl.Scheme {
			case "https", "http":
				fname, err := DownloadToolHttp(cacheDir, archive.U, parsedUrl, archive.H)
				if err != nil {
					return err
				}
				err = ExtractAndLinkTool(tool, fname)
				if err != nil {
					return err
				}

			default:
				slog.Warn("unsupported scheme for tool", "tool", toolName, "scheme", parsedUrl.Scheme)
				continue
			}

		case config.ToolSourceGitRepo:
			slog.Warn("git-repo tool source not implemented")
			continue

		default:
			slog.Warn("invalid tool source", "source", tool.Source)
			continue
		}

		slog.Debug("installed", "tool", toolName)
	}

	return nil
}
