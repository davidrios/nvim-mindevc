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

	"github.com/davidrios/nvim-mindevc/config"
	"github.com/ulikunitz/xz"
)

func DownloadToolHttp(downloadDir string, rawUrl string, parsedUrl *url.URL, expectedHash string) (string, error) {
	cachedFilename := filepath.Join(downloadDir, expectedHash)

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
			slog.Debug("using cached file", "url", rawUrl, "hash", expectedHash)
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

	tmpName := filepath.Join(downloadDir, expectedHash+".tmp")

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

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
		}
	}

	return nil
}

func extractZipFile(f *zip.File, destDir string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	dest := filepath.Join(destDir, f.Name)

	if f.FileInfo().IsDir() {
		err = os.MkdirAll(dest, f.FileInfo().Mode())
		if err != nil {
			return err
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	outFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, rc)
	if err != nil {
		return err
	}

	return nil
}

func extractZip(src, destDir string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if err = extractZipFile(f, destDir); err != nil {
			return err
		}
	}

	return nil
}

func CreateToolSymlinks(extractedTo string, symlinks map[string]string) error {
	for link, target := range symlinks {
		var finalTarget string

		if target == "$bin" {
			finalTarget = extractedTo
			err := os.Chmod(finalTarget, 0o755)
			if err != nil {
				return fmt.Errorf("failed to set executable mode: %w", err)
			}
		} else {
			finalTarget = filepath.Join(extractedTo, target)
		}

		if err := os.MkdirAll(filepath.Dir(link), 0755); err != nil {
			return fmt.Errorf("failed to create directory for symlink %s: %w", link, err)
		}

		if _, err := os.Lstat(link); err == nil {
			if err := os.Remove(link); err != nil {
				return fmt.Errorf("failed to remove existing symlink %s: %w", link, err)
			}
		}

		if err := os.Symlink(finalTarget, link); err != nil {
			return fmt.Errorf("failed to create symlink %s -> %s: %w", link, finalTarget, err)
		}

		slog.Debug("created symlink", "link", link, "target", finalTarget)
	}

	return nil
}

func ExtractTool(
	toolName string,
	archiveType config.ConfigToolArchiveType,
	arch config.ConfigToolArch,
	fname string,
) (string, error) {
	if !archiveType.IsValid() {
		return "", fmt.Errorf("unknown archive type")
	}

	toolDestDir := filepath.Join(filepath.Dir(fname), "..", string(arch), toolName)
	if err := os.MkdirAll(toolDestDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create extraction directory: %w", err)
	}

	var err error

	if archiveType == config.ArchiveTypeZip {
		err = extractZip(fname, toolDestDir)
		if err != nil {
			return "", fmt.Errorf("could not extract zip: %w", err)
		}
	} else {
		toolFile, err := os.Open(fname)
		if err != nil {
			return "", fmt.Errorf("could not open downloaded file: %w", err)
		}
		defer toolFile.Close()

		var toolFileReader io.Reader = toolFile

		switch archiveType {
		case config.ArchiveTypeBin, config.ArchiveTypeBinGz, config.ArchiveTypeBinBz2, config.ArchiveTypeBinXz:
			destFile, err := os.OpenFile(
				filepath.Join(toolDestDir, toolName),
				os.O_CREATE|os.O_RDWR|os.O_TRUNC,
				os.FileMode(0o755))
			if err != nil {
				return "", fmt.Errorf("error writing tool file: %w", err)
			}
			defer destFile.Close()

			switch archiveType {
			case config.ArchiveTypeBinGz:
				toolFileReader, err = gzip.NewReader(toolFile)
			case config.ArchiveTypeBinBz2:
				toolFileReader = bzip2.NewReader(toolFile)
			case config.ArchiveTypeBinXz:
				toolFileReader, err = xz.NewReader(toolFile)
			}

			if err != nil {
				return "", fmt.Errorf("error extracting: %w", err)
			}

			if _, err := io.Copy(destFile, toolFileReader); err != nil {
				return "", fmt.Errorf("error writing tool file: %w", err)
			}

		case config.ArchiveTypeTarGz, config.ArchiveTypeTarBz2, config.ArchiveTypeTarXz:
			switch archiveType {
			case config.ArchiveTypeTarGz:
				toolFileReader, err = gzip.NewReader(toolFile)
			case config.ArchiveTypeTarBz2:
				toolFileReader = bzip2.NewReader(toolFile)
			case config.ArchiveTypeTarXz:
				toolFileReader, err = xz.NewReader(toolFile)
			}

			if err != nil {
				return "", fmt.Errorf("error extracting: %w", err)
			}

			if archiveType.IsTar() {
				err = extractTar(toolFileReader, toolDestDir)
				if err != nil {
					return "", fmt.Errorf("failed to extract tar: %w", err)
				}
			}
		}
	}

	return toolDestDir, nil
}

func GetDownloadsDir(base string) (string, error) {
	downloadDir := filepath.Join(base, "tools", "_download")

	if err := os.MkdirAll(downloadDir, 0o750); err != nil {
		return "", fmt.Errorf("error creating cache dir: %w", err)
	}

	return downloadDir, nil
}

func DownloadTools(
	cacheDir string,
	arch config.ConfigToolArch,
	toolNames []string,
	tools config.ConfigTools,
) ([]string, error) {
	slog.Debug("downloading tools")

	cacheDir, err := config.ExpandHome(cacheDir)
	if err != nil {
		return nil, err
	}

	downloadDir, err := GetDownloadsDir(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("error creating cache dir: %w", err)
	}

	paths := []string{}
	for _, toolName := range toolNames {
		tool, ok := tools[toolName]
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

			parsedUrl, err := url.Parse(archive.Url)
			if err != nil {
				slog.Warn("invalid url for tool", "tool", toolName)
				continue
			}

			switch parsedUrl.Scheme {
			case "https", "http":
				fname, err := DownloadToolHttp(downloadDir, archive.Url, parsedUrl, archive.Hash)
				if err != nil {
					return nil, err
				}
				if toolName == "nvim-mindevc" {
					fname, err = ExtractTool(toolName, archive.Type, arch, fname)
					if err != nil {
						return nil, err
					}
					fname = filepath.Join(fname, toolName)
				}
				paths = append(paths, fname)

			default:
				slog.Warn("unsupported scheme for tool", "tool", toolName, "scheme", parsedUrl.Scheme)
				continue
			}

		case config.ToolSourceGitRepo:
			slog.Warn("git-repo tool source not implemented yet")
			continue

		default:
			slog.Warn("invalid tool source", "source", tool.Source)
			continue
		}

		slog.Debug("downloaded", "tool", toolName)
	}

	return paths, nil
}
