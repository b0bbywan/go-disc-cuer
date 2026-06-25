package musicbrainz

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/b0bbywan/go-disc-cuer/types"
)

// stubServer points mbURL at a local handler for the duration of a test.
func stubServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	prev := mbURL
	mbURL = srv.URL
	t.Cleanup(func() { mbURL = prev })
}

const releaseJSON = `{
	"id": "mb-123",
	"title": "The Album",
	"date": "2020-05-01",
	"artist-credit": [{"name": "The Artist"}],
	"media": [{"tracks": [{"title": "First"}, {"title": "Second"}]}]
}`

func assertSampleRelease(t *testing.T, info *types.DiscInfo) {
	t.Helper()
	if info.ID != "mb-123" || info.Title != "The Album" || info.Artist != "The Artist" || info.ReleaseDate != "2020-05-01" {
		t.Errorf("unexpected disc info: %+v", info)
	}
	if len(info.Tracks) != 2 || info.Tracks[0] != "First" || info.Tracks[1] != "Second" {
		t.Errorf("unexpected tracks: %v", info.Tracks)
	}
}

func TestFetchReleaseByID(t *testing.T) {
	stubServer(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/release/mb-123") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if _, err := w.Write([]byte(releaseJSON)); err != nil {
			t.Errorf("writing response: %v", err)
		}
	})

	info, err := FetchReleaseByID("mb-123")
	if err != nil {
		t.Fatalf("FetchReleaseByID() error = %v", err)
	}
	assertSampleRelease(t, info)
}

func TestFetchReleaseByToc(t *testing.T) {
	stubServer(t, func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`{"releases": [` + releaseJSON + `]}`)); err != nil {
			t.Errorf("writing response: %v", err)
		}
	})

	info, err := FetchReleaseByToc("1+2+3")
	if err != nil {
		t.Fatalf("FetchReleaseByToc() error = %v", err)
	}
	assertSampleRelease(t, info)
}

func TestFetchReleaseByTocNoMatch(t *testing.T) {
	stubServer(t, func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`{"releases": []}`)); err != nil {
			t.Errorf("writing response: %v", err)
		}
	})

	if _, err := FetchReleaseByToc("1+2+3"); err == nil {
		t.Fatal("expected error when no releases are returned, got nil")
	}
}

func TestFetchJSONStatusError(t *testing.T) {
	stubServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})

	if _, err := FetchReleaseByID("mb-123"); err == nil {
		t.Fatal("expected error on non-200 status, got nil")
	}
}

func TestConvertReleaseToDiscInfo(t *testing.T) {
	release := types.MBRelease{
		ID:    "id-1",
		Title: "Title",
		Date:  "1999",
	}
	release.ArtistCredit = append(release.ArtistCredit, struct{ Name string }{Name: "Artist"})
	release.Media = append(release.Media, struct {
		Tracks []struct{ Title string }
	}{Tracks: []struct{ Title string }{{Title: "Only"}}})

	info, err := convertReleaseToDiscInfo(release)
	if err != nil {
		t.Fatalf("convertReleaseToDiscInfo() error = %v", err)
	}
	if info.Artist != "Artist" || len(info.Tracks) != 1 || info.Tracks[0] != "Only" {
		t.Errorf("unexpected disc info: %+v", info)
	}
}
