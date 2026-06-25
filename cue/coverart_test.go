package cue

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/b0bbywan/go-disc-cuer/types"
	"github.com/b0bbywan/go-disc-cuer/utils"
)

// coverStub points coverArtURL at a local handler for the duration of a test.
func coverStub(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	prev := coverArtURL
	coverArtURL = srv.URL
	t.Cleanup(func() { coverArtURL = prev })
}

func TestFetchCoverArt(t *testing.T) {
	coverStub(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mb-123/front" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if _, err := w.Write([]byte("JPEGDATA")); err != nil {
			t.Errorf("writing response: %v", err)
		}
	})

	dst := filepath.Join(t.TempDir(), "cover.jpg")
	if err := fetchCoverArt("mb-123", dst); err != nil {
		t.Fatalf("fetchCoverArt() error = %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("reading cover: %v", err)
	}
	if string(got) != "JPEGDATA" {
		t.Errorf("cover content = %q, want %q", got, "JPEGDATA")
	}
}

func TestFetchCoverArtStatusError(t *testing.T) {
	coverStub(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	})

	dst := filepath.Join(t.TempDir(), "cover.jpg")
	if err := fetchCoverArt("mb-123", dst); err == nil {
		t.Fatal("expected error on 404, got nil")
	}
	if _, err := os.Stat(dst); err == nil {
		t.Error("no file should be written on failure")
	}
}

func TestEnsureCoverArtSkipsWhenPresent(t *testing.T) {
	called := false
	coverStub(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	info := &types.DiscInfo{ID: "mb-123", CoverArtPath: "/already/set.jpg"}
	ensureCoverArt(t.TempDir(), info, "disc-1")

	if called {
		t.Error("ensureCoverArt should not hit the network when cover art is already set")
	}
	if info.CoverArtPath != "/already/set.jpg" {
		t.Errorf("CoverArtPath changed to %q", info.CoverArtPath)
	}
}

func TestEnsureCoverArtSetsPathOnSuccess(t *testing.T) {
	coverStub(t, func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("JPEGDATA")); err != nil {
			t.Errorf("writing response: %v", err)
		}
	})

	cache := t.TempDir()
	id := "disc-1"
	// write() creates this folder in the real flow; do the same here.
	if err := utils.CreateFolderIfNeeded(utils.CacheCoverArtPath(cache, id)); err != nil {
		t.Fatalf("seeding cache dir: %v", err)
	}

	info := &types.DiscInfo{ID: "mb-123"}
	ensureCoverArt(cache, info, id)

	want := utils.CacheCoverArtPath(cache, id)
	if info.CoverArtPath != want {
		t.Errorf("CoverArtPath = %q, want %q", info.CoverArtPath, want)
	}
	if _, err := os.Stat(want); err != nil {
		t.Errorf("expected cover file at %s: %v", want, err)
	}
}

func TestEnsureCoverArtNonFatalOnFailure(t *testing.T) {
	coverStub(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	})

	info := &types.DiscInfo{ID: "mb-123"}
	ensureCoverArt(t.TempDir(), info, "disc-1")

	// Failure is logged, not propagated; the path stays empty.
	if info.CoverArtPath != "" {
		t.Errorf("CoverArtPath should stay empty on failure, got %q", info.CoverArtPath)
	}
}
