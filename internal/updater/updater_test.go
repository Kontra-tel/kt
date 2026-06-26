package updater_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.kontra.tel/kontra.tel/build-tools/internal/updater"
)

func releaseServer(t *testing.T, releases []updater.Release) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/releases" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(releases)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestLatestRelease(t *testing.T) {
	assets := []updater.Asset{
		{Name: "kt-linux-amd64", BrowserDownloadURL: "http://example.com/kt-linux-amd64"},
	}
	srv := releaseServer(t, []updater.Release{{TagName: "v2.1.0", Assets: assets}})

	r, err := updater.LatestRelease(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if r.TagName != "v2.1.0" {
		t.Errorf("TagName = %q, want %q", r.TagName, "v2.1.0")
	}
	if len(r.Assets) != 1 || r.Assets[0].Name != "kt-linux-amd64" {
		t.Errorf("Assets not parsed correctly: %+v", r.Assets)
	}
}

func TestLatestRelease_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	_, err := updater.LatestRelease(srv.URL)
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestCheck_Newer(t *testing.T) {
	srv := releaseServer(t, []updater.Release{{TagName: "v1.5.0"}})

	latest, newer, err := updater.Check(srv.URL, "1.0.0", false)
	if err != nil {
		t.Fatal(err)
	}
	if !newer {
		t.Error("expected newer=true for 1.5.0 > 1.0.0")
	}
	if latest != "1.5.0" {
		t.Errorf("latest = %q, want %q", latest, "1.5.0")
	}
}

func TestCheck_UpToDate(t *testing.T) {
	srv := releaseServer(t, []updater.Release{{TagName: "v1.0.0"}})

	_, newer, err := updater.Check(srv.URL, "1.0.0", false)
	if err != nil {
		t.Fatal(err)
	}
	if newer {
		t.Error("expected newer=false when current matches latest")
	}
}

func TestCheck_StripsvPrefix(t *testing.T) {
	srv := releaseServer(t, []updater.Release{{TagName: "v2.0.0"}})

	latest, _, err := updater.Check(srv.URL, "1.0.0", false)
	if err != nil {
		t.Fatal(err)
	}
	if latest != "2.0.0" {
		t.Errorf("latest = %q, expected v prefix stripped", latest)
	}
}

func TestCheck_NetworkError(t *testing.T) {
	_, _, err := updater.Check("http://127.0.0.1:0", "1.0.0", false)
	if err == nil {
		t.Fatal("expected error for unreachable host, got nil")
	}
}

func TestCheck_IgnoresPrereleaseByDefault(t *testing.T) {
	srv := releaseServer(t, []updater.Release{
		{TagName: "v2.0.0-rc.1", Prerelease: true},
		{TagName: "v1.9.9"},
	})
	latest, newer, err := updater.Check(srv.URL, "1.9.8", false)
	if err != nil {
		t.Fatal(err)
	}
	if latest != "1.9.9" || !newer {
		t.Fatalf("latest=%q newer=%v", latest, newer)
	}
}

func TestCheck_CanSelectPrerelease(t *testing.T) {
	srv := releaseServer(t, []updater.Release{
		{TagName: "v2.0.0-rc.1", Prerelease: true},
		{TagName: "v1.9.9"},
	})
	latest, newer, err := updater.Check(srv.URL, "1.9.9", true)
	if err != nil {
		t.Fatal(err)
	}
	if latest != "2.0.0-rc.1" || !newer {
		t.Fatalf("latest=%q newer=%v", latest, newer)
	}
}
