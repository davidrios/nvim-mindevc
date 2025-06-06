package setup

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadToolHttp_Success(t *testing.T) {
	dir, err := os.MkdirTemp("", "gotest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	testContent := "test file content"
	expectedHash := fmt.Sprintf("%x", sha256.Sum256([]byte(testContent)))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(testContent))
	}))
	defer ts.Close()

	filename := "testfile.tar.gz"
	burl := fmt.Sprintf("%s/%s", ts.URL, filename)
	parsedUrl, _ := url.Parse(burl)
	err = DownloadToolHttp(dir, burl, parsedUrl, expectedHash)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expectedPath := filepath.Join(dir, filename)

	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("Expected file to be created at %s", expectedPath)
	}

	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != testContent {
		t.Fatalf("Expected content %q, got %q", testContent, string(content))
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
	err = DownloadToolHttp(dir, burl, parsedUrl, "somehash")
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
	err = DownloadToolHttp(dir, ts.URL, parsedUrl, "somehash")
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
	err = DownloadToolHttp(dir, ts.URL, parsedUrl, wrongHash)
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
	err := DownloadToolHttp(invalidDir, ts.URL, parsedUrl, "somehash")
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
	err = DownloadToolHttp(dir, ts.URL, parsedUrl, "")
	if err == nil {
		t.Fatalf("Expected error for empty response")
	}
}
