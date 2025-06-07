package setup

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/davidrios/nvim-mindevc/config"
	"github.com/ulikunitz/xz"
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

func ungz(r io.Reader) (io.Reader, error) {
	return gzip.NewReader(r)
}

func unxz(r io.Reader) (io.Reader, error) {
	return xz.NewReader(r)
}

func unbz2(r io.Reader) io.Reader {
	return bzip2.NewReader(r)
}

func extractTar(r io.Reader, dest string) error {
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}

	return nil
}

func extractTarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := ungz(file)
	if err != nil {
		return err
	}
	defer gzr.(io.Closer).Close()

	return extractTar(gzr, dest)
}

func extractTarXz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	xzr, err := unxz(file)
	if err != nil {
		return err
	}

	return extractTar(xzr, dest)
}

func extractTarBz2(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	bzr := unbz2(file)
	return extractTar(bzr, dest)
}

func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.FileInfo().Mode())
			rc.Close()
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			rc.Close()
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func handleBin(src, dest, binName string) error {
	target := filepath.Join(dest, binName)
	
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

func createSymlinks(extractDir string, symlinks map[string]string) error {
	for linkPath, target := range symlinks {
		var finalTarget string
		
		if target == "$bin" {
			finalTarget = extractDir
		} else {
			finalTarget = filepath.Join(extractDir, target)
		}

		if err := os.MkdirAll(filepath.Dir(linkPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for symlink %s: %w", linkPath, err)
		}

		if _, err := os.Lstat(linkPath); err == nil {
			if err := os.Remove(linkPath); err != nil {
				return fmt.Errorf("failed to remove existing symlink %s: %w", linkPath, err)
			}
		}

		if err := os.Symlink(finalTarget, linkPath); err != nil {
			return fmt.Errorf("failed to create symlink %s -> %s: %w", linkPath, finalTarget, err)
		}

		slog.Debug("created symlink", "link", linkPath, "target", finalTarget)
	}

	return nil
}

func ExtractAndLinkTool(tool config.ConfigTool, archive config.ConfigToolArchive, fname string) error {
	var extractDir string
	
	switch archive.T {
	case config.ArchiveTypeTarGz, config.ArchiveTypeTarXz, config.ArchiveTypeTarBz2, config.ArchiveTypeZip:
		extractDir = strings.TrimSuffix(fname, filepath.Ext(fname))
		if err := os.MkdirAll(extractDir, 0755); err != nil {
			return fmt.Errorf("failed to create extraction directory: %w", err)
		}
		
		switch archive.T {
		case config.ArchiveTypeTarGz:
			if err := extractTarGz(fname, extractDir); err != nil {
				return fmt.Errorf("failed to extract tar.gz: %w", err)
			}
		case config.ArchiveTypeTarXz:
			if err := extractTarXz(fname, extractDir); err != nil {
				return fmt.Errorf("failed to extract tar.xz: %w", err)
			}
		case config.ArchiveTypeTarBz2:
			if err := extractTarBz2(fname, extractDir); err != nil {
				return fmt.Errorf("failed to extract tar.bz2: %w", err)
			}
		case config.ArchiveTypeZip:
			if err := extractZip(fname, extractDir); err != nil {
				return fmt.Errorf("failed to extract zip: %w", err)
			}
		}
		
	case config.ArchiveTypeBin:
		extractDir = fname
		
	default:
		return fmt.Errorf("unsupported archive type: %s", archive.T)
	}

	if err := createSymlinks(extractDir, tool.Symlinks); err != nil {
		return fmt.Errorf("failed to create symlinks: %w", err)
	}

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
				err = ExtractAndLinkTool(tool, archive, fname)
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
