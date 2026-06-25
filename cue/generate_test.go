package cue

import (
	"errors"
	"os"
	"strings"
	"testing"

	"go.uploadedlobster.com/discid"

	"github.com/b0bbywan/go-disc-cuer/config"
	"github.com/b0bbywan/go-disc-cuer/types"
	"github.com/b0bbywan/go-disc-cuer/utils"
)

// issue9MBTOC is the MusicBrainz TOC of the disc from issue #9; discid.Parse
// turns it into a real disc so the whole flow runs without a drive.
const (
	issue9MBTOC  = "1 21 347590 150 15947 29333 47188 62249 81995 99079 118776 131854 146350 163525 182410 195995 209445 227574 243154 257085 272920 292952 308907 327349"
	issue9DiscID = "23121815"
)

// parsedDiscOpener builds a disc opener that parses a TOC string instead of
// reading physical hardware.
func parsedDiscOpener(toc string) Option {
	return withDiscOpener(func(string) (readableDisc, error) {
		return discid.Parse(toc)
	})
}

func TestGenerateEndToEnd(t *testing.T) {
	t.Parallel()
	cache := t.TempDir()
	cfg := &config.Config{CacheLocation: cache}

	var gotData discData
	resolver := withResolver(func(_ Options, d discData) (*types.DiscInfo, error) {
		gotData = d
		return &types.DiscInfo{
			Artist:       "Various",
			Title:        "NRJ Hits 2003",
			ReleaseDate:  "2003",
			Genre:        "Pop",
			Tracks:       []string{"Hey Ho", "Papi Chulo"},
			CoverArtPath: "/cover.jpg", // non-empty: skips the cover-art network call
		}, nil
	})

	g := New(cfg, parsedDiscOpener(issue9MBTOC), resolver)
	path, err := g.Generate(Options{})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// The disc id derived from the real TOC pipeline drives the cache path.
	want := utils.CachePlaylistPath(cache, issue9DiscID)
	if path != want {
		t.Fatalf("Generate() = %q, want %q", path, want)
	}

	// The resolver received the TOCs computed from the parsed disc.
	if gotData.id != issue9DiscID {
		t.Errorf("resolver disc id = %q, want %q", gotData.id, issue9DiscID)
	}
	if !strings.HasPrefix(gotData.gnuToc, issue9DiscID+" ") {
		t.Errorf("resolver gnu toc = %q, want prefix %q", gotData.gnuToc, issue9DiscID)
	}
	if gotData.mbToc != issue9MBTOC {
		t.Errorf("resolver mb toc = %q, want %q", gotData.mbToc, issue9MBTOC)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading sheet: %v", err)
	}
	for _, want := range []string{`TITLE "NRJ Hits 2003"`, `PERFORMER "Various"`, `TITLE "Hey Ho"`} {
		if !strings.Contains(string(content), want) {
			t.Errorf("sheet missing %q:\n%s", want, content)
		}
	}
}

func TestGenerateOverwrite(t *testing.T) {
	t.Parallel()
	cache := t.TempDir()
	cfg := &config.Config{CacheLocation: cache}

	calls := 0
	g := New(cfg, parsedDiscOpener(issue9MBTOC), withResolver(func(_ Options, _ discData) (*types.DiscInfo, error) {
		calls++
		return &types.DiscInfo{Artist: "A", Title: "T", Tracks: []string{"x"}, CoverArtPath: "/c.jpg"}, nil
	}))

	// First run populates the cache.
	if _, err := g.Generate(Options{}); err != nil {
		t.Fatalf("first Generate() error = %v", err)
	}
	// Without Overwrite the cached sheet is returned without resolving again.
	if _, err := g.Generate(Options{}); err != nil {
		t.Fatalf("cached Generate() error = %v", err)
	}
	if calls != 1 {
		t.Fatalf("resolver called %d times, want 1 (cache hit on second run)", calls)
	}
	// With Overwrite it regenerates.
	if _, err := g.Generate(Options{Overwrite: true}); err != nil {
		t.Fatalf("overwrite Generate() error = %v", err)
	}
	if calls != 2 {
		t.Fatalf("resolver called %d times, want 2 (overwrite forces resolve)", calls)
	}
}

func TestGenerateOpenDiscError(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{CacheLocation: t.TempDir()}
	g := New(cfg, withDiscOpener(func(string) (readableDisc, error) {
		return nil, errors.New("no drive")
	}))

	if _, err := g.Generate(Options{}); err == nil {
		t.Fatal("expected error when the drive cannot be opened, got nil")
	}
}
