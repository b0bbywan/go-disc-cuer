package cue

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/b0bbywan/go-disc-cuer/config"
	"github.com/b0bbywan/go-disc-cuer/types"
	"github.com/b0bbywan/go-disc-cuer/utils"
)

func TestOptionsValidate(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
	}{
		{name: "empty is valid", opts: Options{}},
		{name: "musicbrainz alone is valid", opts: Options{MusicBrainzID: "mb-1"}},
		{name: "disc id with musicbrainz is valid", opts: Options{DiscID: "abc", MusicBrainzID: "mb-1"}},
		{name: "disc id without musicbrainz fails", opts: Options{DiscID: "abc"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestRenderCue(t *testing.T) {
	info := &types.DiscInfo{
		Artist:       "The Artist",
		Title:        "The Album",
		ReleaseDate:  "2024-01-01",
		Genre:        "Rock",
		CoverArtPath: "/cache/cover.jpg",
		Tracks:       []string{"First", "Second"},
	}

	got := string(renderCue(info))

	want := strings.Join([]string{
		`REM DATE "2024-01-01"`,
		`REM GENRE "Rock"`,
		`REM COVER "/cache/cover.jpg"`,
		`PERFORMER "The Artist"`,
		`TITLE "The Album"`,
		`FILE "cdda:///1" WAVE`,
		`  TRACK 01 AUDIO`,
		`    TITLE "First"`,
		`FILE "cdda:///2" WAVE`,
		`  TRACK 02 AUDIO`,
		`    TITLE "Second"`,
		"",
	}, "\n")

	if got != want {
		t.Errorf("renderCue() mismatch:\n got:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderCueOmitsEmptyOptionalFields(t *testing.T) {
	info := &types.DiscInfo{
		Artist: "A",
		Title:  "T",
		Tracks: []string{"Only"},
	}

	got := string(renderCue(info))

	for _, rem := range []string{"REM DATE", "REM GENRE", "REM COVER"} {
		if strings.Contains(got, rem) {
			t.Errorf("renderCue() should omit %q when unset, got:\n%s", rem, got)
		}
	}
	want := strings.Join([]string{
		`PERFORMER "A"`,
		`TITLE "T"`,
		`FILE "cdda:///1" WAVE`,
		`  TRACK 01 AUDIO`,
		`    TITLE "Only"`,
		"",
	}, "\n")
	if got != want {
		t.Errorf("renderCue() mismatch:\n got:\n%s\nwant:\n%s", got, want)
	}
}

func TestWriteCreatesSheet(t *testing.T) {
	dir := t.TempDir()
	id := "discid-1"
	path := utils.CachePlaylistPath(dir, id)
	info := &types.DiscInfo{
		Artist:       "A",
		Title:        "T",
		Tracks:       []string{"One"},
		CoverArtPath: "/already/set.jpg", // non-empty: skips network cover fetch
	}

	got, err := write(dir, info, id, path)
	if err != nil {
		t.Fatalf("write() error = %v", err)
	}
	if got != path {
		t.Fatalf("write() returned %q, want %q", got, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading written sheet: %v", err)
	}
	if !strings.Contains(string(content), `PERFORMER "A"`) {
		t.Errorf("written sheet missing performer line:\n%s", content)
	}
}

func TestGenerateNilConfig(t *testing.T) {
	if _, err := New(nil).Generate(Options{}); err == nil {
		t.Fatal("Generate with nil config expected error, got nil")
	}
}

func TestGenerateValidationError(t *testing.T) {
	cfg := &config.Config{CacheLocation: t.TempDir()}
	if _, err := New(cfg).Generate(Options{DiscID: "abc"}); err == nil {
		t.Fatal("Generate with disc id but no musicbrainz id expected error, got nil")
	}
}

func TestGenerateReturnsCachedSheet(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{CacheLocation: dir}
	id := "cached-disc"
	path := utils.CachePlaylistPath(dir, id)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("seeding cache dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("cached"), 0o644); err != nil {
		t.Fatalf("seeding cache file: %v", err)
	}

	// DiscID set => no drive read; cached file present and Overwrite false => returned as-is.
	got, err := New(cfg).Generate(Options{DiscID: id, MusicBrainzID: "mb-1"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if got != path {
		t.Fatalf("Generate() = %q, want cached %q", got, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading sheet: %v", err)
	}
	if string(content) != "cached" {
		t.Errorf("cached sheet was overwritten, content = %q", content)
	}
}
