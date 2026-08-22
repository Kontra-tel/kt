package updater

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"git.kontra.tel/kontra.tel/Kt/internal/versioning"
)

type Release struct {
	TagName    string  `json:"tag_name"`
	Prerelease bool    `json:"prerelease"`
	Assets     []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// LatestRelease fetches the latest stable release from the Gitea API.
func LatestRelease(apiBase string) (Release, error) {
	return SelectRelease(apiBase, false)
}

func ListReleases(apiBase string) ([]Release, error) {
	resp, err := http.Get(apiBase + "/releases")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned %s", resp.Status)
	}
	var releases []Release
	return releases, json.NewDecoder(resp.Body).Decode(&releases)
}

func SelectRelease(apiBase string, includePrerelease bool) (Release, error) {
	releases, err := ListReleases(apiBase)
	if err != nil {
		return Release{}, err
	}
	var chosen Release
	var chosenVersion versioning.Version
	found := false
	for _, r := range releases {
		if r.Prerelease && !includePrerelease {
			continue
		}
		v, err := versioning.Parse(strings.TrimPrefix(r.TagName, "v"))
		if err != nil {
			continue
		}
		if !found || v.Compare(chosenVersion) > 0 {
			chosen = r
			chosenVersion = v
			found = true
		}
	}
	if !found {
		if includePrerelease {
			return Release{}, fmt.Errorf("no releases found")
		}
		return Release{}, fmt.Errorf("no stable releases found")
	}
	return chosen, nil
}

// Check returns the latest version and whether it is newer than current.
func Check(apiBase, current string, includePrerelease bool) (string, bool, error) {
	r, err := SelectRelease(apiBase, includePrerelease)
	if err != nil {
		return "", false, err
	}
	latest := strings.TrimPrefix(r.TagName, "v")
	cur, err := versioning.Parse(current)
	if err != nil {
		return "", false, err
	}
	lat, err := versioning.Parse(latest)
	if err != nil {
		return "", false, err
	}
	return latest, lat.Compare(cur) > 0, nil
}

// Apply downloads the selected release binary for the current OS/arch and
// atomically replaces the running executable.
func Apply(apiBase string, includePrerelease bool) error {
	r, err := SelectRelease(apiBase, includePrerelease)
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

	tmp, err := os.CreateTemp(filepath.Dir(exe), "kt-update-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

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
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := verifyChecksum(r, want, tmpPath); err != nil {
		return err
	}

	return os.Rename(tmpPath, exe)
}

func verifyChecksum(r Release, assetName, path string) error {
	var sumsURL string
	for _, a := range r.Assets {
		if a.Name == "SHA256SUMS" {
			sumsURL = a.BrowserDownloadURL
			break
		}
	}
	if sumsURL == "" {
		return fmt.Errorf("release %s has no SHA256SUMS asset", r.TagName)
	}
	resp, err := http.Get(sumsURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("checksum download returned %s", resp.Status)
	}
	want := ""
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[1] == assetName {
			want = fields[0]
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if want == "" {
		return fmt.Errorf("SHA256SUMS has no checksum for %s", assetName)
	}
	got, err := fileSHA256(path)
	if err != nil {
		return err
	}
	if !strings.EqualFold(got, want) {
		return fmt.Errorf("checksum mismatch for %s", assetName)
	}
	return nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ExecutablePath returns the real path of the running binary, resolving symlinks.
func ExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}
