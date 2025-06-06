package setup

import (
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
			err = DownloadToolHttp(dir, burl, parsedUrl, tv.hash)
			if err != nil {
				if tv.hashFail && err.Error() == "hashes do not match" {
					return
				}

				t.Fatalf("Expected no error, got: %v", err)
			}

			expectedPath := filepath.Join(dir, tv.fname)

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
