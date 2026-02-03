package cue

import "github.com/b0bbywan/go-disc-cuer/config"

// GenerationOptions holds all options for CUE file generation
type GenerationOptions struct {
	// Device is the CD-ROM device path (e.g., "/dev/sr0")
	Device string

	// Config contains application configuration
	Config *config.Config

	// ProvidedDiscID is an optional user-supplied disc ID
	ProvidedDiscID string

	// MusicBrainzID is an optional MusicBrainz release ID for metadata
	MusicBrainzID string

	// Overwrite forces regeneration even if CUE file exists
	Overwrite bool
}

// DiscData holds disc-specific data (IDs and TOCs)
type DiscData struct {
	// DiscID is the FreeDB/GNUDB disc identifier
	DiscID string

	// GnuTOC is the table of contents formatted for GNUDB
	GnuTOC string

	// MusicBrainzTOC is the table of contents formatted for MusicBrainz
	MusicBrainzTOC string
}
