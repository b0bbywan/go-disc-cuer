package utils

import "testing"

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
