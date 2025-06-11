package setup

import (
	"crypto/sha256"
	"fmt"
	"io"
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

func CheckFileHash(t *testing.T, fpath string, hash string) {
	t.Helper()
	fp, err := os.Open(fpath)
	if err != nil {
		t.Fatalf("error reading file: %s", err)
	}
	defer fp.Close()

	h := sha256.New()
	if _, err := io.Copy(h, fp); err != nil {
		t.Fatalf("error creating hash for: %s", fpath)
	}

	gotHash := fmt.Sprintf("%x", h.Sum(nil))

	if gotHash != hash {
		t.Fatalf("file hash doesn't match")
	}
}

func TestExtractAndLinkTool(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-extract-link")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	downloadDir := filepath.Join(tempDir, "_download")
	if err = os.MkdirAll(downloadDir, 0o750); err != nil {
		t.Fatalf("%s", err)
	}

	files := map[string]string{
		"a":           "87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7",
		"aa/aa":       "d9cd8155764c3543f10fad8a480d743137466f8d55213c8eaefcd12f06d43a80",
		"bb/bb":       "a81c31ac62620b9215a14ff00544cb07a55b765594f3ab3be77e70923ae27cf1",
		"bb/cc/dd/dd": "b6f9dd313cde39ae1b87e63b9b457029bcea6e9520b5db5de20d3284e4c0259e",
	}
	links := map[string]string{
		"_links/aax": "aa/aa",
		"_links/ddx": "bb/cc/dd/dd",
	}

	testTable := []struct {
		name        string
		archiveType config.ConfigToolArchiveType
		fname       string
		prefix      string
	}{{
		name:        "tool1",
		archiveType: config.ArchiveTypeZip,
		fname:       "tool.zip",
	}, {
		name:        "tool2",
		archiveType: config.ArchiveTypeTarGz,
		fname:       "tool.tar.gz",
		prefix:      "extracted-v1.0.0",
	}, {
		name:        "tool3",
		archiveType: config.ArchiveTypeTarBz2,
		fname:       "tool.tar.bz2",
		prefix:      "extracted-v1.0.0",
	}, {
		name:        "tool4",
		archiveType: config.ArchiveTypeTarXz,
		fname:       "tool.tar.xz",
		prefix:      "extracted-v1.0.0",
	}, {
		name:        "tool5",
		archiveType: config.ArchiveTypeBin,
		fname:       "rg",
	}, {
		name:        "tool6",
		archiveType: config.ArchiveTypeBinGz,
		fname:       "rg.gz",
	}, {
		name:        "tool7",
		archiveType: config.ArchiveTypeBinBz2,
		fname:       "rg.bz2",
	}, {
		name:        "tool8",
		archiveType: config.ArchiveTypeBinXz,
		fname:       "rg.xz",
	},
	}
	for _, tv := range testTable {
		t.Run(tv.name, func(t *testing.T) {
			downloadFname := filepath.Join(downloadDir, tv.fname)

			fp, err := os.OpenFile(downloadFname, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			defer fp.Close()

			sfp, err := os.Open(filepath.Join("testdata", "extractfiles", tv.fname))
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			defer sfp.Close()

			if _, err := io.Copy(fp, sfp); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			extracted, err := ExtractTool(tv.name, tv.archiveType, "", downloadFname)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			switch tv.archiveType {
			case config.ArchiveTypeBin, config.ArchiveTypeBinGz, config.ArchiveTypeBinBz2, config.ArchiveTypeBinXz:
				CheckFileHash(t, filepath.Join(extracted, tv.prefix, tv.name), "3f212a63b283e660406da7b022b52be26ea74a893c41ce093b00b2e5b0b36d5c")
			case config.ArchiveTypeTarGz, config.ArchiveTypeTarBz2, config.ArchiveTypeTarXz:
				for fname, hash := range files {
					CheckFileHash(t, filepath.Join(extracted, tv.prefix, fname), hash)
				}
			}

			err = os.MkdirAll(filepath.Join(tempDir, "_links"), 0o755)
			if err != nil {
				t.Fatalf("error: %s", err)
			}

			switch tv.archiveType {
			case config.ArchiveTypeBin, config.ArchiveTypeBinGz, config.ArchiveTypeBinBz2, config.ArchiveTypeBinXz:
				link := filepath.Join(tempDir, "_links", tv.name)
				err = CreateToolSymlinks(extracted, map[string]string{link: "$bin"})
				if err != nil {
					t.Fatalf("error: %s", err)
				}
				_, err := os.Stat(link)
				if err != nil {
					t.Fatalf("error: %s", err)
				}
			case config.ArchiveTypeTarGz, config.ArchiveTypeTarBz2, config.ArchiveTypeTarXz:
				linksWithPrefix := map[string]string{}
				for link, target := range links {
					linksWithPrefix[filepath.Join(tempDir, link)] = filepath.Join(tv.prefix, target)
				}

				err = CreateToolSymlinks(extracted, linksWithPrefix)
				if err != nil {
					t.Fatalf("error: %s", err)
				}

				for link := range linksWithPrefix {
					_, err := os.Stat(link)
					if err != nil {
						t.Fatalf("error: %s", err)
					}
				}
			}
		})
	}
}
