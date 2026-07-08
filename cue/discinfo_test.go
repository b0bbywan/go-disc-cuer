package cue

import (
	"errors"
	"testing"

	"github.com/b0bbywan/go-disc-cuer/types"
)

func TestSelectDiscInfo(t *testing.T) {
	gnudb := &types.DiscInfo{ID: "gnudb-id", Artist: "GNUDB Artist", Title: "GNUDB Title"}
	mb := &types.DiscInfo{ID: "mb-id", Artist: "MB Artist", Title: "MB Title"}

	t.Run("both fail returns error", func(t *testing.T) {
		info, err := selectDiscInfo(nil, errors.New("gnudb down"), nil, errors.New("mb down"))
		if err == nil {
			t.Fatal("expected error when both sources fail")
		}
		if info != nil {
			t.Errorf("expected nil info on dual failure, got %+v", info)
		}
	})

	t.Run("gnudb preferred for body, mb id kept", func(t *testing.T) {
		info, err := selectDiscInfo(gnudb, nil, mb, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Artist != "GNUDB Artist" || info.Title != "GNUDB Title" {
			t.Errorf("expected GNUDB body, got artist=%q title=%q", info.Artist, info.Title)
		}
		if info.ID != "mb-id" {
			t.Errorf("expected MusicBrainz id %q for cover lookup, got %q", "mb-id", info.ID)
		}
	})

	t.Run("falls back to musicbrainz when gnudb fails", func(t *testing.T) {
		info, err := selectDiscInfo(nil, errors.New("gnudb down"), mb, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Artist != "MB Artist" || info.ID != "mb-id" {
			t.Errorf("expected MusicBrainz fallback, got %+v", info)
		}
	})

	t.Run("uses gnudb when musicbrainz fails", func(t *testing.T) {
		info, err := selectDiscInfo(gnudb, nil, nil, errors.New("mb down"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Artist != "GNUDB Artist" {
			t.Errorf("expected GNUDB body, got %+v", info)
		}
		// MusicBrainz failed, so its id must not clobber the existing one.
		if info.ID != "gnudb-id" {
			t.Errorf("expected gnudb id retained when mb fails, got %q", info.ID)
		}
	})
}
