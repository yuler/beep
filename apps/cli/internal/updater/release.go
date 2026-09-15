package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"beep/internal/version"
)

const (
	DefaultRepo = "yuler/beep"
)

// GetRepo returns the repository to check, honoring BEEP_REPO override if set.
func GetRepo() string {
	if r := os.Getenv("BEEP_REPO"); r != "" {
		return strings.TrimSpace(r)
	}
	return DefaultRepo
}

// ReleaseAsset represents an attached binary archive or checksum file in a release.
type ReleaseAsset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
	Size        int64  `json:"size"`
}

// ReleaseInfo holds the metadata of a GitHub Release.
type ReleaseInfo struct {
	Tag         string         `json:"tag_name"`
	Name        string         `json:"name"`
	HTMLURL     string         `json:"html_url"`
	PublishedAt time.Time      `json:"published_at"`
	Assets      []ReleaseAsset `json:"assets"`
}

// apiHTTPClient is used for short GitHub API / metadata requests.
var apiHTTPClient = &http.Client{
	Timeout: 15 * time.Second,
}

// downloadHTTPClient is used for large asset downloads. It has no Client.Timeout
// so long transfers are not cut off; callers must cancel via context instead.
var downloadHTTPClient = &http.Client{}

// SetHTTPClientsForTest replaces the package HTTP clients for tests and returns
// a restore function that puts the originals back.
func SetHTTPClientsForTest(api, download *http.Client) (restore func()) {
	prevAPI, prevDownload := apiHTTPClient, downloadHTTPClient
	if api != nil {
		apiHTTPClient = api
	}
	if download != nil {
		downloadHTTPClient = download
	}
	return func() {
		apiHTTPClient = prevAPI
		downloadHTTPClient = prevDownload
	}
}

// FetchLatestRelease fetches the latest published release for the repository.
func FetchLatestRelease(ctx context.Context, repo string) (*ReleaseInfo, error) {
	if repo == "" {
		repo = GetRepo()
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", fmt.Sprintf("beep-cli/%s", version.Version))

	resp, err := apiHTTPClient.Do(req)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var rel ReleaseInfo
			if err := json.NewDecoder(resp.Body).Decode(&rel); err == nil && rel.Tag != "" {
				return &rel, nil
			}
		}
	}

	// Fallback to GitHub Web redirect resolution (e.g. if API is rate-limited)
	redirectTag, err := resolveLatestTagRedirect(ctx, repo)
	if err == nil && redirectTag != "" {
		return &ReleaseInfo{
			Tag:     redirectTag,
			HTMLURL: fmt.Sprintf("https://github.com/%s/releases/tag/%s", repo, redirectTag),
		}, nil
	}

	return nil, fmt.Errorf("failed to fetch latest release for %s from GitHub", repo)
}

// FetchReleaseByTag fetches a specific release by git tag.
func FetchReleaseByTag(ctx context.Context, repo, tag string) (*ReleaseInfo, error) {
	if repo == "" {
		repo = GetRepo()
	}
	tag = strings.TrimSpace(tag)
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", repo, tag)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", fmt.Sprintf("beep-cli/%s", version.Version))

	resp, err := apiHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error fetching release %s: %w", tag, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release tag %s not found on %s", tag, repo)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("failed to fetch release %s: status %d (%s)", tag, resp.StatusCode, string(body))
	}

	var rel ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("invalid release JSON: %w", err)
	}
	return &rel, nil
}

// resolveLatestTagRedirect resolves the latest tag via HTTP redirect headers from github.com/{repo}/releases/latest.
func resolveLatestTagRedirect(ctx context.Context, repo string) (string, error) {
	webURL := fmt.Sprintf("https://github.com/%s/releases/latest", repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, webURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("beep-cli/%s", version.Version))

	noRedirectClient := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	loc := resp.Header.Get("Location")
	if loc == "" {
		return "", fmt.Errorf("no Location header in redirect response")
	}

	idx := strings.LastIndex(loc, "/")
	if idx == -1 || idx == len(loc)-1 {
		return "", fmt.Errorf("invalid redirect location: %s", loc)
	}
	tag := loc[idx+1:]
	if tag == "latest" || tag == "" {
		return "", fmt.Errorf("could not parse tag from redirect %s", loc)
	}
	return tag, nil
}
