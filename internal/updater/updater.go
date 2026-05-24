package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// LatestRelease fetches the latest release from the Gitea API.
func LatestRelease(apiBase string) (Release, error) {
	resp, err := http.Get(apiBase + "/releases/latest")
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("API returned %s", resp.Status)
	}
	var r Release
	return r, json.NewDecoder(resp.Body).Decode(&r)
}

// Check returns the latest version and whether it is newer than current.
// Returns (latest, false, nil) when already up to date.
func Check(apiBase, current string) (string, bool, error) {
	r, err := LatestRelease(apiBase)
	if err != nil {
		return "", false, err
	}
	latest := strings.TrimPrefix(r.TagName, "v")
	return latest, latest != "" && latest != current, nil
}

// Apply downloads the latest release binary for the current OS/arch and
// atomically replaces the running executable.
func Apply(apiBase string) error {
	r, err := LatestRelease(apiBase)
	if err != nil {
		return err
	}

	want := fmt.Sprintf("kt-%s-%s", runtime.GOOS, runtime.GOARCH)
	var downloadURL string
	for _, a := range r.Assets {
		if a.Name == want {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("no release asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	exe, err := ExecutablePath()
	if err != nil {
		return err
	}

	// Download to a temp file in the same directory so os.Rename is atomic.
	tmp, err := os.CreateTemp(filepath.Dir(exe), "kt-update-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpPath)
	}()

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned %s", resp.Status)
	}

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return err
	}
	if err := tmp.Chmod(0755); err != nil {
		return err
	}
	tmp.Close()

	return os.Rename(tmpPath, exe)
}

// ExecutablePath returns the real path of the running binary, resolving symlinks.
func ExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}
