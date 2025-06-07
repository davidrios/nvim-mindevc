package setup

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidrios/nvim-mindevc/config"
)

func TestDownloadToolHttp_Success(t *testing.T) {
	dir, err := os.MkdirTemp("", "gotest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	testTable := []struct {
		fname    string
		content  string
		hash     string
		hashFail bool
	}{
		{
			fname:   "testfile.tar.gz",
			content: "test content",
			hash:    "6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72"},
		{
			fname:   "testfile2.tar.gz",
			content: "another content",
			hash:    "6292c8b17c54333d0449794f91ca2287c29e0adc1bcf06795c54bc6aa1a003e6"},
		{
			fname:    "testfile3.tar.gz",
			content:  "another content",
			hash:     "5292c8b17c54333d0449794f91ca2287c29e0adc1bcf06795c54bc6aa1a003e6",
			hashFail: true},
	}

	for _, tv := range testTable {
		t.Run(tv.fname, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tv.content))
			}))
			defer ts.Close()

			burl := fmt.Sprintf("%s/%s", ts.URL, tv.fname)
			parsedUrl, _ := url.Parse(burl)
			_, err = DownloadToolHttp(dir, burl, parsedUrl, tv.hash)
			if err != nil {
				if tv.hashFail && err.Error() == "hashes do not match" {
					return
				}

				t.Fatalf("Expected no error, got: %v", err)
			}

			expectedPath := filepath.Join(dir, tv.hash)

			if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
				t.Fatalf("Expected file to be created at %s", expectedPath)
			}

			content, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("Failed to read downloaded file: %v", err)
			}

			if string(content) != tv.content {
				t.Fatalf("Expected content %q, got %q", tv.content, string(content))
			}
		})
	}
}

func TestDownloadToolHttp_InvalidURL(t *testing.T) {
	dir, err := os.MkdirTemp("", "gotest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	burl := "invalid-url"
	parsedUrl, _ := url.Parse(burl)
	_, err = DownloadToolHttp(dir, burl, parsedUrl, "somehash")
	if err == nil {
		t.Fatal("Expected error for invalid URL, got nil")
	}
}

func TestDownloadToolHttp_HTTPError(t *testing.T) {
	dir, err := os.MkdirTemp("", "gotest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	parsedUrl, _ := url.Parse(ts.URL)
	_, err = DownloadToolHttp(dir, ts.URL, parsedUrl, "somehash")
	if err == nil {
		t.Fatal("Expected error for 404 response, got nil")
	}
}

func TestDownloadToolHttp_HashMismatch(t *testing.T) {
	dir, err := os.MkdirTemp("", "gotest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	testContent := "test file content"
	wrongHash := "wronghash"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(testContent))
	}))
	defer ts.Close()

	parsedUrl, _ := url.Parse(ts.URL)
	_, err = DownloadToolHttp(dir, ts.URL, parsedUrl, wrongHash)
	if err == nil {
		t.Fatal("Expected error for hash mismatch, got nil")
	}
}

func TestDownloadToolHttp_InvalidCacheDir(t *testing.T) {
	invalidDir := "/nonexistent/path/that/should/not/exist"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	}))
	defer ts.Close()

	parsedUrl, _ := url.Parse(ts.URL)
	_, err := DownloadToolHttp(invalidDir, ts.URL, parsedUrl, "somehash")
	if err == nil {
		t.Fatal("Expected error for invalid cache directory, got nil")
	}
}

func TestDownloadToolHttp_EmptyResponse(t *testing.T) {
	dir, err := os.MkdirTemp("", "gotest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	parsedUrl, _ := url.Parse(ts.URL)
	_, err = DownloadToolHttp(dir, ts.URL, parsedUrl, "")
	if err == nil {
		t.Fatalf("Expected error for empty response")
	}
}

func TestDownloadToolHttp_CachingBehavior(t *testing.T) {
	dir, err := os.MkdirTemp("", "gotest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	testContent := "test content for caching"
	expectedHash := "47f30c7ebec2e17fd2ff5bb93dcfc189773e5d90fd1dbf8f9dbed877973174e3"
	requestCount := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(testContent))
	}))
	defer ts.Close()

	parsedUrl, _ := url.Parse(ts.URL)

	// First download
	fname1, err := DownloadToolHttp(dir, ts.URL, parsedUrl, expectedHash)
	if err != nil {
		t.Fatalf("First download failed: %v", err)
	}

	if requestCount != 1 {
		t.Fatalf("Expected 1 HTTP request, got %d", requestCount)
	}

	// Verify file exists with hash name
	expectedPath := filepath.Join(dir, expectedHash)
	if fname1 != expectedPath {
		t.Fatalf("Expected filename %s, got %s", expectedPath, fname1)
	}

	// Second download should use cache
	fname2, err := DownloadToolHttp(dir, ts.URL, parsedUrl, expectedHash)
	if err != nil {
		t.Fatalf("Second download failed: %v", err)
	}

	if requestCount != 1 {
		t.Fatalf("Expected 1 HTTP request (cached), got %d", requestCount)
	}

	if fname1 != fname2 {
		t.Fatalf("Expected same filename for cached file, got %s vs %s", fname1, fname2)
	}

	// Verify content is correct
	content, err := os.ReadFile(fname2)
	if err != nil {
		t.Fatalf("Failed to read cached file: %v", err)
	}

	if string(content) != testContent {
		t.Fatalf("Expected content %q, got %q", testContent, string(content))
	}
}

func createTestTarGz(t *testing.T, dest string, files map[string]string) {
	file, err := os.Create(dest)
	if err != nil {
		t.Fatalf("Failed to create tar.gz file: %v", err)
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for name, content := range files {
		header := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatalf("Failed to write tar header: %v", err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("Failed to write tar content: %v", err)
		}
	}
}

func createTestZip(t *testing.T, dest string, files map[string]string) {
	file, err := os.Create(dest)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	defer zw.Close()

	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("Failed to create zip entry: %v", err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("Failed to write zip content: %v", err)
		}
	}
}

func TestExtractTarGz(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-extract")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFiles := map[string]string{
		"tool/bin/fd":     "fake fd binary",
		"tool/doc/README": "documentation",
	}

	tarFile := filepath.Join(tempDir, "test.tar.gz")
	createTestTarGz(t, tarFile, testFiles)

	extractDir := filepath.Join(tempDir, "extracted")
	if err := extractTarGz(tarFile, extractDir); err != nil {
		t.Fatalf("Failed to extract tar.gz: %v", err)
	}

	for name, expectedContent := range testFiles {
		extractedPath := filepath.Join(extractDir, name)
		content, err := os.ReadFile(extractedPath)
		if err != nil {
			t.Fatalf("Failed to read extracted file %s: %v", name, err)
		}
		if string(content) != expectedContent {
			t.Fatalf("Content mismatch for %s: expected %q, got %q", name, expectedContent, string(content))
		}
	}
}

func TestExtractZip(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-extract")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFiles := map[string]string{
		"tool/bin/rg":     "fake ripgrep binary",
		"tool/doc/README": "documentation",
	}

	zipFile := filepath.Join(tempDir, "test.zip")
	createTestZip(t, zipFile, testFiles)

	extractDir := filepath.Join(tempDir, "extracted")
	if err := extractZip(zipFile, extractDir); err != nil {
		t.Fatalf("Failed to extract zip: %v", err)
	}

	for name, expectedContent := range testFiles {
		extractedPath := filepath.Join(extractDir, name)
		content, err := os.ReadFile(extractedPath)
		if err != nil {
			t.Fatalf("Failed to read extracted file %s: %v", name, err)
		}
		if string(content) != expectedContent {
			t.Fatalf("Content mismatch for %s: expected %q, got %q", name, expectedContent, string(content))
		}
	}
}

func TestExtractAndLinkTool_TarGz(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-extract-link")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFiles := map[string]string{
		"fd-v10.2.0-x86_64-unknown-linux-musl/fd": "fake fd binary",
	}

	tarFile := filepath.Join(tempDir, "test.tar.gz")
	createTestTarGz(t, tarFile, testFiles)

	tool := config.ConfigTool{
		Symlinks: map[string]string{
			filepath.Join(tempDir, "usr/local/bin/fd"): "fd-v10.2.0-x86_64-unknown-linux-musl/fd",
		},
	}

	archive := config.ConfigToolArchive{
		T: config.ArchiveTypeTarGz,
	}

	if err := ExtractAndLinkTool(tool, archive, tarFile); err != nil {
		t.Fatalf("Failed to extract and link tool: %v", err)
	}

	linkPath := filepath.Join(tempDir, "usr/local/bin/fd")
	if _, err := os.Lstat(linkPath); err != nil {
		t.Fatalf("Expected symlink to be created at %s: %v", linkPath, err)
	}
}

func TestExtractAndLinkTool_Bin(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-extract-bin")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	binContent := "fake gosu binary"
	binFile := filepath.Join(tempDir, "gosu-amd64")
	if err := os.WriteFile(binFile, []byte(binContent), 0755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	tool := config.ConfigTool{
		Symlinks: map[string]string{
			filepath.Join(tempDir, "usr/local/bin/gosu"): "$bin",
		},
	}

	archive := config.ConfigToolArchive{
		T: config.ArchiveTypeBin,
	}

	if err := ExtractAndLinkTool(tool, archive, binFile); err != nil {
		t.Fatalf("Failed to extract and link binary tool: %v", err)
	}

	linkPath := filepath.Join(tempDir, "usr/local/bin/gosu")
	if _, err := os.Lstat(linkPath); err != nil {
		t.Fatalf("Expected symlink to be created at %s: %v", linkPath, err)
	}

	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("Failed to read symlink target: %v", err)
	}

	if target != binFile {
		t.Fatalf("Expected symlink target %s, got %s", binFile, target)
	}
}
