package setup

import (
	"compress/bzip2"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"

	"github.com/davidrios/nvim-mindevc/config"
	"github.com/davidrios/nvim-mindevc/utils"
)

func DownloadToolHttp(downloadDir string, rawUrl string, parsedUrl *url.URL, expectedHash string) (string, error) {
	cachedFilename := filepath.Join(downloadDir, expectedHash)
	slog.Debug("download cache name", "n", cachedFilename)

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

	tmpName := cachedFilename + ".tmp"
	err := utils.DownloadFileHttp(rawUrl, tmpName)
	if err != nil {
		return "", err
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

func CreateToolSymlinks(extractedTo string, symlinks map[string]string) error {
	for link, target := range symlinks {
		var finalTarget string

		if target == "$bin" {
			finalTarget = filepath.Join(extractedTo, filepath.Base(extractedTo))
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
		err = utils.ExtractZip(fname, toolDestDir)
		if err != nil {
			return "", fmt.Errorf("could not extract zip: %w", err)
		}
	} else {
		uncFile, err := UncompressTool(archiveType, fname)
		if err != nil {
			return "", err
		}
		var toolFile *os.File
		if uncFile == "" {
			toolFile, err = os.Open(fname)
		} else {
			toolFile, err = os.Open(uncFile)
		}
		if err != nil {
			return "", fmt.Errorf("could not open downloaded file: %w", err)
		}
		defer toolFile.Close()

		switch archiveType {
		case config.ArchiveTypeBin, config.ArchiveTypeBinGz, config.ArchiveTypeBinBz2, config.ArchiveTypeBinXz:
			destFile, err := os.Create(filepath.Join(toolDestDir, toolName))
			if err != nil {
				return "", fmt.Errorf("error writing tool file: %w", err)
			}
			defer destFile.Close()

			if _, err := io.Copy(destFile, toolFile); err != nil {
				return "", fmt.Errorf("error writing tool file: %w", err)
			}

		case config.ArchiveTypeTarGz, config.ArchiveTypeTarBz2, config.ArchiveTypeTarXz:
			slog.Debug("extracting tar")
			err = utils.ExtractTar(toolFile, toolDestDir)
			if err != nil {
				return "", fmt.Errorf("failed to extract tar: %w", err)
			}
		}
	}

	return toolDestDir, nil
}

func UncompressTool(
	archiveType config.ConfigToolArchiveType,
	fname string,
) (string, error) {
	if !archiveType.IsValid() {
		return "", fmt.Errorf("unknown archive type")
	}

	if archiveType == config.ArchiveTypeZip || !archiveType.IsGBXZCompressed() {
		return "", nil
	}

	toolFile, err := os.Open(fname)
	if err != nil {
		return "", fmt.Errorf("could not open downloaded file: %w", err)
	}
	defer toolFile.Close()

	var toolFileReader io.Reader = toolFile

	uncFile := fname + ".unc"

	if _, err := os.Stat(uncFile); err != nil {
		err = nil
		switch archiveType {
		case config.ArchiveTypeTarGz, config.ArchiveTypeBinGz:
			toolFileReader, err = gzip.NewReader(toolFile)
		case config.ArchiveTypeTarBz2, config.ArchiveTypeBinBz2:
			toolFileReader = bzip2.NewReader(toolFile)
		case config.ArchiveTypeTarXz, config.ArchiveTypeBinXz:
			toolFileReader, err = xz.NewReader(toolFile)
		}

		if err != nil {
			return "", fmt.Errorf("error extracting: %w", err)
		}

		uncTmp, err := os.Create(uncFile + ".tmp")

		if err != nil {
			return "", err
		}

		if _, err := io.Copy(uncTmp, toolFileReader); err != nil {
			uncTmp.Close()
			return "", err
		}
		uncTmp.Close()

		err = os.Rename(uncFile+".tmp", uncFile)
		if err != nil {
			return "", err
		}
	}

	return uncFile, nil
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
) (map[string]string, error) {
	slog.Debug("downloading tools")

	downloadDir, err := GetDownloadsDir(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("error creating cache dir: %w", err)
	}

	paths := make(map[string]string)
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
					if toolName == "nvim-mindevc" {
						fname = filepath.Join(fname, toolName)
					}
				}
				paths[toolName] = fname

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

func ExtractTools(
	arch config.ConfigToolArch,
	toolNames []string,
	tools config.ConfigTools,
	downloaded map[string]string,
) (map[string]string, error) {
	slog.Debug("extracting tools")

	paths := make(map[string]string)
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
			path, err := ExtractTool(toolName, archive.Type, arch, downloaded[toolName])
			if err != nil {
				return nil, err
			}

			paths[toolName] = path

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

func LinkTools(
	arch config.ConfigToolArch,
	toolNames []string,
	tools config.ConfigTools,
	extracted map[string]string,
) error {
	slog.Debug("linking tools")

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
			err := CreateToolSymlinks(extracted[toolName], archive.Links)
			if err != nil {
				return err
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

	return nil
}

func DownloadAndExtractLocalTools(cacheDir string) error {
	cmd := exec.Command("uname", "-sm")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("could not get current OS: %w", err)
	}

	osArch := strings.TrimSpace(string(output))
	if !strings.Contains(osArch, "Linux") {
		return fmt.Errorf("Error: TODO, implement other OSes support")
	}
	arch := config.ConfigToolArch(osArch[strings.Index(osArch, " ")+1:])

	locDownloaded, err := DownloadTools(cacheDir,
		config.ConfigToolArch(arch),
		[]string{"zig"},
		config.ConfigTools{"zig": config.ZigTool},
	)
	if err != nil {
		return err
	}

	extractedTo, err := ExtractTool("zig", config.ZigTool.Archives[arch].Type, arch, locDownloaded["zig"])
	if err != nil {
		return err
	}

	if err = CreateToolSymlinks(extractedTo, map[string]string{
		filepath.Join(cacheDir, "bin", "zig"): config.ZigTool.Archives[arch].Links["/usr/local/bin/zig"],
	}); err != nil {
		return err
	}

	return nil
}
