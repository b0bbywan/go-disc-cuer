package utils

import (
	"testing"

	"go.uploadedlobster.com/discid"
)

// issue9 is the disc from https://github.com/b0bbywan/go-disc-cuer/issues/9,
// parsed from its MusicBrainz TOC so the real TOC pipeline runs without a drive.
const (
	issue9MBTOC   = "1 21 347590 150 15947 29333 47188 62249 81995 99079 118776 131854 146350 163525 182410 195995 209445 227574 243154 257085 272920 292952 308907 327349"
	issue9DiscID  = "23121815"
	issue9GnuTOC  = "23121815 21 150 15947 29333 47188 62249 81995 99079 118776 131854 146350 163525 182410 195995 209445 227574 243154 257085 272920 292952 308907 327349 4634"
)

func TestGetTocAndDiscIDFromParsedDisc(t *testing.T) {
	disc, err := discid.Parse(issue9MBTOC)
	if err != nil {
		t.Fatalf("parsing TOC: %v", err)
	}
	defer disc.Close()

	gnuToc, id, err := GetTocAndDiscID(disc)
	if err != nil {
		t.Fatalf("GetTocAndDiscID() error = %v", err)
	}
	if id != issue9DiscID {
		t.Errorf("disc id = %q, want %q", id, issue9DiscID)
	}
	if gnuToc != issue9GnuTOC {
		t.Errorf("gnu toc =\n%q\nwant\n%q", gnuToc, issue9GnuTOC)
	}
}

func TestGetMusicBrainzTOCFromParsedDisc(t *testing.T) {
	disc, err := discid.Parse(issue9MBTOC)
	if err != nil {
		t.Fatalf("parsing TOC: %v", err)
	}
	defer disc.Close()

	got, err := GetMusicBrainzTOC(disc)
	if err != nil {
		t.Fatalf("GetMusicBrainzTOC() error = %v", err)
	}
	if got != issue9MBTOC {
		t.Errorf("mb toc = %q, want %q", got, issue9MBTOC)
	}
}

func TestGetTocAndDiscIDNilDisc(t *testing.T) {
	if _, _, err := GetTocAndDiscID(nil); err == nil {
		t.Fatal("GetTocAndDiscID(nil) expected error, got nil")
	}
}

func TestGetMusicBrainzTOCNilDisc(t *testing.T) {
	if _, err := GetMusicBrainzTOC(nil); err == nil {
		t.Fatal("GetMusicBrainzTOC(nil) expected error, got nil")
	}
}

func TestTocToGnuNilDisc(t *testing.T) {
	if _, err := tocToGnu(nil); err == nil {
		t.Fatal("tocToGnu(nil) expected error, got nil")
	}
}
