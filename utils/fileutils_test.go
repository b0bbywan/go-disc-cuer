package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCachePaths(t *testing.T) {
	cache := "/var/cache/disc-cuer"
	id := "abc123"

	if got, want := CachePlaylistPath(cache, id), filepath.Join(cache, id, "playlist.cue"); got != want {
		t.Errorf("CachePlaylistPath = %q, want %q", got, want)
	}
	if got, want := CacheCoverArtPath(cache, id), filepath.Join(cache, id, "cover.jpg"); got != want {
		t.Errorf("CacheCoverArtPath = %q, want %q", got, want)
	}
}

func TestCheckIfPlaylistExists(t *testing.T) {
	dir := t.TempDir()

	missing := filepath.Join(dir, "nope.cue")
	if CheckIfPlaylistExists(missing) {
		t.Error("CheckIfPlaylistExists returned true for a missing file")
	}

	present := filepath.Join(dir, "there.cue")
	if err := os.WriteFile(present, []byte("x"), 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	if !CheckIfPlaylistExists(present) {
		t.Error("CheckIfPlaylistExists returned false for an existing file")
	}
}

func TestCreateFolderIfNeeded(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "sub", "deep", "playlist.cue")

	if err := CreateFolderIfNeeded(target); err != nil {
		t.Fatalf("CreateFolderIfNeeded error = %v", err)
	}

	info, err := os.Stat(filepath.Dir(target))
	if err != nil {
		t.Fatalf("expected created dir to exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected a directory to be created")
	}
}
