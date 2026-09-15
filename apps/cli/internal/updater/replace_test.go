package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractTarGz(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.tar.gz")

	// Create a test tar.gz containing "beep"
	expectedContent := "#!/bin/sh\necho test\n"
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name:     "beep",
		Mode:     0755,
		Size:     int64(len(expectedContent)),
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	if _, err := tw.Write([]byte(expectedContent)); err != nil {
		t.Fatalf("tw.Write: %v", err)
	}
	tw.Close()
	gw.Close()

	if err := os.WriteFile(archivePath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	var extracted bytes.Buffer
	if err := extractTarGz(archivePath, "beep", &extracted); err != nil {
		t.Fatalf("extractTarGz failed: %v", err)
	}

	if extracted.String() != expectedContent {
		t.Errorf("extracted content = %q; want %q", extracted.String(), expectedContent)
	}
}

func TestExtractZip(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.zip")

	expectedContent := "binary content here"
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("beep.exe")
	if err != nil {
		t.Fatalf("zw.Create: %v", err)
	}
	if _, err := w.Write([]byte(expectedContent)); err != nil {
		t.Fatalf("w.Write: %v", err)
	}
	zw.Close()

	if err := os.WriteFile(archivePath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	var extracted bytes.Buffer
	if err := extractZip(archivePath, "beep.exe", &extracted); err != nil {
		t.Fatalf("extractZip failed: %v", err)
	}

	if extracted.String() != expectedContent {
		t.Errorf("extracted content = %q; want %q", extracted.String(), expectedContent)
	}
}

func TestFetchExpectedChecksum(t *testing.T) {
	checksumsContent := `
e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  beep_0.2.2_darwin_arm64.tar.gz
a1b2c3d4e5f67890123456789abcdef0123456789abcdef0123456789abcdef0 *beep_0.2.2_linux_amd64.tar.gz
1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef  beep_0.2.2_windows_amd64.zip
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(checksumsContent))
	}))
	defer server.Close()

	hash, err := fetchExpectedChecksum(context.Background(), server.URL, "beep_0.2.2_linux_amd64.tar.gz")
	if err != nil {
		t.Fatalf("fetchExpectedChecksum failed: %v", err)
	}
	expected := "a1b2c3d4e5f67890123456789abcdef0123456789abcdef0123456789abcdef0"
	if hash != expected {
		t.Errorf("got hash %q, want %q", hash, expected)
	}

	// Missing entry
	_, err = fetchExpectedChecksum(context.Background(), server.URL, "nonexistent.tar.gz")
	if err == nil {
		t.Errorf("expected error for nonexistent file, got nil")
	}
}

func TestCheckIfHomebrew(t *testing.T) {
	cases := []struct {
		path     string
		expected bool
	}{
		{"/opt/homebrew/bin/beep", true},
		{"/usr/local/Cellar/beep/0.2.1/bin/beep", true},
		{"/home/linuxbrew/.linuxbrew/bin/beep", true},
		{"/usr/local/bin/beep", false},
		{"/home/yule/.local/bin/beep", false},
	}

	for _, c := range cases {
		got := CheckIfHomebrew(c.path)
		if got != c.expected {
			t.Errorf("CheckIfHomebrew(%q) = %v; want %v", c.path, got, c.expected)
		}
	}
}
