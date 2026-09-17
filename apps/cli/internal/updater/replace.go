package updater

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"beep/internal/config"
	"beep/internal/version"
)

// maxArchiveBytes caps the size of a downloaded release archive (256 MiB).
const maxArchiveBytes int64 = 256 << 20

// UpgradeOptions specifies parameters for upgrading beep CLI.
type UpgradeOptions struct {
	TargetVersion string // "latest" or specific tag like "v0.2.2"
	Force         bool
	Repo          string // default "yuler/beep"
	Workspace     string
	OnProgress    func(stage string)
	// Release, when set with a non-empty Tag, skips FetchLatestRelease / FetchReleaseByTag.
	Release *ReleaseInfo
}

// UpgradeResult contains details of a completed upgrade.
type UpgradeResult struct {
	OldVersion     string
	NewVersion     string
	ExecutablePath string
	InstalledTag   string
}

// CheckIfHomebrew checks if the executable was installed via Homebrew.
func CheckIfHomebrew(path string) bool {
	lower := strings.ToLower(path)
	return strings.Contains(lower, "/cellar/") ||
		strings.Contains(lower, "/homebrew/") ||
		strings.Contains(lower, "/.linuxbrew/")
}

// Upgrade downloads and replaces the current beep binary with the target release.
func Upgrade(ctx context.Context, opts UpgradeOptions) (*UpgradeResult, error) {
	repo := opts.Repo
	if repo == "" {
		repo = GetRepo()
	}

	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to determine executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve symlink for executable %s: %w", execPath, err)
	}

	// Homebrew check
	if CheckIfHomebrew(execPath) && !opts.Force {
		return nil, fmt.Errorf("%s appears to be installed via Homebrew (%s).\nWe recommend running: 'brew upgrade beep'\nOr pass '--force' to proceed with binary upgrade anyway", config.BinaryName(), execPath)
	}

	targetDir := filepath.Dir(execPath)

	// Check writable permissions on target directory
	probeFile := filepath.Join(targetDir, fmt.Sprintf(".beep-probe-%d.tmp", time.Now().UnixNano()))
	if err := os.WriteFile(probeFile, []byte("ok"), 0o600); err != nil {
		return nil, fmt.Errorf("permission denied writing to %s\nFix directory ownership/permissions, or if installed via Homebrew run: brew upgrade beep", targetDir)
	}
	_ = os.Remove(probeFile)

	if opts.OnProgress != nil {
		opts.OnProgress("Resolving release version...")
	}

	var rel *ReleaseInfo
	if opts.Release != nil && opts.Release.Tag != "" {
		rel = opts.Release
	} else {
		targetVer := strings.TrimSpace(opts.TargetVersion)
		if targetVer == "" || strings.EqualFold(targetVer, "latest") {
			rel, err = FetchLatestRelease(ctx, repo)
		} else {
			rel, err = FetchReleaseByTag(ctx, repo, targetVer)
		}
		if err != nil {
			return nil, err
		}
	}

	rawVersion := rel.Tag
	if rawVersion == "" {
		return nil, fmt.Errorf("could not determine release version from GitHub")
	}

	// Compare with current version
	if !opts.Force {
		cmp := CompareVersions(rawVersion, version.Version)
		if cmp == 0 {
			return nil, fmt.Errorf("%s is already at the latest version (%s)", config.BinaryName(), version.Version)
		}
	}

	// Platform file naming matching install.sh
	vClean := strings.TrimPrefix(rawVersion, "v")
	osName := runtime.GOOS
	archName := runtime.GOARCH

	ext := "tar.gz"
	binFileName := "beep"
	if osName == "windows" {
		ext = "zip"
		binFileName = "beep.exe"
	}

	archiveName := fmt.Sprintf("beep_%s_%s_%s.%s", vClean, osName, archName, ext)
	archiveURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repo, rawVersion, archiveName)
	checksumURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/checksums.txt", repo, rawVersion)

	if opts.OnProgress != nil {
		opts.OnProgress(fmt.Sprintf("Downloading %s (%s)...", archiveName, rawVersion))
	}

	// Create temp archive in targetDir to avoid cross-device moves
	archiveTmp, err := os.CreateTemp(targetDir, ".beep-archive-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary download file: %w", err)
	}
	archiveTmpPath := archiveTmp.Name()
	defer os.Remove(archiveTmpPath)

	hasher := sha256.New()
	multiWriter := io.MultiWriter(archiveTmp, hasher)

	downloadReq, err := http.NewRequestWithContext(ctx, http.MethodGet, archiveURL, nil)
	if err != nil {
		archiveTmp.Close()
		return nil, err
	}
	downloadReq.Header.Set("User-Agent", fmt.Sprintf("beep-cli/%s", version.Version))

	downloadResp, err := downloadHTTPClient.Do(downloadReq)
	if err != nil {
		archiveTmp.Close()
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer downloadResp.Body.Close()

	if downloadResp.StatusCode != http.StatusOK {
		archiveTmp.Close()
		return nil, fmt.Errorf("failed to download release asset from %s (HTTP %d)", archiveURL, downloadResp.StatusCode)
	}

	if err := copyArchiveLimited(multiWriter, downloadResp.Body, downloadResp.ContentLength, maxArchiveBytes); err != nil {
		archiveTmp.Close()
		return nil, fmt.Errorf("error writing archive: %w", err)
	}
	archiveTmp.Close()

	actualHash := hex.EncodeToString(hasher.Sum(nil))

	if opts.OnProgress != nil {
		opts.OnProgress("Verifying checksums.txt...")
	}

	expectedHash, err := fetchExpectedChecksum(ctx, checksumURL, archiveName)
	if err != nil {
		return nil, fmt.Errorf("checksum verification failed: %w", err)
	}

	if !strings.EqualFold(actualHash, expectedHash) {
		return nil, fmt.Errorf("checksum mismatch for %s:\n  Expected: %s\n  Actual:   %s", archiveName, expectedHash, actualHash)
	}

	if opts.OnProgress != nil {
		opts.OnProgress("Extracting binary...")
	}

	newBinTmp, err := os.CreateTemp(targetDir, ".beep-bin-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp binary: %w", err)
	}
	newBinPath := newBinTmp.Name()
	defer os.Remove(newBinPath)

	if ext == "tar.gz" {
		if err := extractTarGz(archiveTmpPath, binFileName, newBinTmp); err != nil {
			newBinTmp.Close()
			return nil, fmt.Errorf("extraction error: %w", err)
		}
	} else if ext == "zip" {
		if err := extractZip(archiveTmpPath, binFileName, newBinTmp); err != nil {
			newBinTmp.Close()
			return nil, fmt.Errorf("extraction error: %w", err)
		}
	}
	newBinTmp.Close()

	if err := os.Chmod(newBinPath, 0o755); err != nil {
		return nil, fmt.Errorf("failed to set binary permissions: %w", err)
	}

	if opts.OnProgress != nil {
		opts.OnProgress(fmt.Sprintf("Replacing %s...", execPath))
	}

	// Atomic binary replacement
	if runtime.GOOS == "windows" {
		oldExePath := execPath + ".old"
		_ = os.Remove(oldExePath)
		if err := os.Rename(execPath, oldExePath); err != nil {
			return nil, fmt.Errorf("failed to rename existing Windows binary: %w", err)
		}
		if err := os.Rename(newBinPath, execPath); err != nil {
			_ = os.Rename(oldExePath, execPath)
			return nil, fmt.Errorf("failed to replace Windows binary: %w", err)
		}
		_ = os.Remove(oldExePath)
	} else {
		if err := os.Rename(newBinPath, execPath); err != nil {
			return nil, fmt.Errorf("failed to replace binary %s: %w", execPath, err)
		}
	}

	// Update state cache so probe does not notify for this version
	ws := opts.Workspace
	if ws == "" {
		ws = config.DefaultWorkspace()
	}
	statePath := GetStateFilePath(ws)
	st, _ := LoadState(statePath)
	if st == nil {
		st = &State{}
	}
	st.LastCheckedAt = time.Now()
	st.LatestVersion = rawVersion
	st.NotifiedVersion = rawVersion
	_ = SaveState(statePath, st)

	return &UpgradeResult{
		OldVersion:     version.Version,
		NewVersion:     rawVersion,
		ExecutablePath: execPath,
		InstalledTag:   rawVersion,
	}, nil
}

// copyArchiveLimited copies from src to dst, rejecting bodies larger than maxBytes.
// When contentLength > 0 and exceeds maxBytes, it fails early without reading.
func copyArchiveLimited(dst io.Writer, src io.Reader, contentLength, maxBytes int64) error {
	if contentLength > 0 && contentLength > maxBytes {
		return fmt.Errorf("release archive Content-Length %d exceeds limit of %d bytes", contentLength, maxBytes)
	}

	limited := io.LimitReader(src, maxBytes+1)
	n, err := io.Copy(dst, limited)
	if err != nil {
		return err
	}
	if n > maxBytes {
		return fmt.Errorf("release archive exceeds size limit of %d bytes", maxBytes)
	}
	return nil
}

func fetchExpectedChecksum(ctx context.Context, checksumURL, archiveName string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, checksumURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("beep-cli/%s", version.Version))

	resp, err := apiHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch checksums.txt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksums.txt returned HTTP %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			hash := fields[0]
			name := strings.TrimPrefix(fields[1], "*")
			if name == archiveName {
				return hash, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading checksums.txt: %w", err)
	}

	return "", fmt.Errorf("archive %s not found in checksums.txt", archiveName)
}

func extractTarGz(archivePath, targetFilename string, dest io.Writer) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzReader, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		baseName := filepath.Base(header.Name)
		if baseName == targetFilename && header.Typeflag == tar.TypeReg {
			_, err = io.Copy(dest, tarReader)
			return err
		}
	}

	return fmt.Errorf("binary %q not found in tar.gz archive", targetFilename)
}

func extractZip(archivePath, targetFilename string, dest io.Writer) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if filepath.Base(f.Name) == targetFilename {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()
			_, err = io.Copy(dest, rc)
			return err
		}
	}

	return fmt.Errorf("binary %q not found in zip archive", targetFilename)
}
