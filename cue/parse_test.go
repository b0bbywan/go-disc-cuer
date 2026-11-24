package cue

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/b0bbywan/go-disc-cuer/types"
)

func TestParseCueRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		info *types.DiscInfo
	}{
		{
			name: "all fields",
			info: &types.DiscInfo{
				Artist:       "Bénabar",
				Title:        "Reprise des négociations",
				ReleaseDate:  "2005",
				Genre:        "Chanson",
				CoverArtPath: "/cache/9c0bf40c/cover.jpg",
				Tracks:       []string{"Le dîner", "Maritie et Gilbert Carpentier"},
			},
		},
		{
			name: "optional fields missing",
			info: &types.DiscInfo{Artist: "A", Title: "T", Tracks: []string{"Only"}},
		},
		{
			name: "quotes in values",
			info: &types.DiscInfo{
				Artist: `The "Band"`,
				Title:  `Say "Hello"`,
				Tracks: []string{`"Quoted"`, `A "B" C`},
			},
		},
		{
			name: "empty titles",
			info: &types.DiscInfo{Tracks: []string{"", "Second", ""}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCue(renderCue(tt.info))
			if !reflect.DeepEqual(got, tt.info) {
				t.Errorf("parseCue(renderCue()) = %+v, want %+v", got, tt.info)
			}
		})
	}
}

func TestParseCueSkipsUnknownLines(t *testing.T) {
	data := []byte("REM COMMENT \"ExactAudioCopy\"\r\nPERFORMER \"A\"\r\nTITLE \"T\"\r\n" +
		"FILE \"cdda:///1\" WAVE\r\n  TRACK 01 AUDIO\r\n    PERFORMER \"Guest\"\r\n    INDEX 01 00:00:00\r\n")

	want := &types.DiscInfo{Artist: "A", Title: "T", Tracks: []string{""}}
	if got := parseCue(data); !reflect.DeepEqual(got, want) {
		t.Errorf("parseCue() = %+v, want %+v", got, want)
	}
}

func TestRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "playlist.cue")
	want := &types.DiscInfo{Artist: "A", Title: "T", Tracks: []string{"One", "Two"}}
	if err := os.WriteFile(path, renderCue(want), 0o644); err != nil {
		t.Fatalf("seeding sheet: %v", err)
	}

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Read() = %+v, want %+v", got, want)
	}
}

func TestReadMissingFile(t *testing.T) {
	if _, err := Read(filepath.Join(t.TempDir(), "missing.cue")); err == nil {
		t.Fatal("Read() of a missing sheet expected error, got nil")
	}
}
