package cue

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"go.uploadedlobster.com/discid"

	"github.com/b0bbywan/go-disc-cuer/config"
	"github.com/b0bbywan/go-disc-cuer/musicbrainz"
	"github.com/b0bbywan/go-disc-cuer/types"
	"github.com/b0bbywan/go-disc-cuer/utils"
)

// Options controls a single CUE generation.
type Options struct {
	Device        string // drive to read when no DiscID is supplied
	DiscID        string // FreeDB disc id, requires MusicBrainzID
	MusicBrainzID string // release id, bypasses the TOC lookup
	Overwrite     bool   // regenerate even if a cached sheet exists
}

func (o Options) validate() error {
	if o.DiscID != "" && o.MusicBrainzID == "" {
		return errors.New("disc id requires a musicbrainz id")
	}
	return nil
}

// discData is what a single drive read yields.
type discData struct {
	id     string
	gnuToc string
	mbToc  string
}

// Generate resolves disc metadata and writes a CUE sheet to the cache, returning
// its path. A cached sheet is reused unless Overwrite is set. Metadata comes from
// MusicBrainzID when given, otherwise from a concurrent GNUDB/MusicBrainz lookup by
// TOC. The drive is read only when no DiscID is supplied.
func Generate(cfg *config.Config, opts Options) (string, error) {
	if cfg == nil {
		return "", errors.New("nil config")
	}
	if err := opts.validate(); err != nil {
		return "", err
	}

	disc := discData{id: opts.DiscID}
	if disc.id == "" {
		var err error
		if disc, err = readDisc(opts.Device); err != nil {
			return "", err
		}
	}

	path := utils.CachePlaylistPath(cfg.GetCacheLocation(), disc.id)
	if !opts.Overwrite && utils.CheckIfPlaylistExists(path) {
		return path, nil
	}

	info, err := resolveInfo(cfg, opts, disc)
	if err != nil {
		return "", err
	}
	return write(cfg.GetCacheLocation(), info, disc.id, path)
}

// readDisc reads the drive once and computes the cache id and both TOCs.
func readDisc(device string) (discData, error) {
	disc, err := discid.Read(device)
	if err != nil {
		return discData{}, err
	}
	defer disc.Close()

	gnuToc, id, err := utils.GetTocAndDiscID(disc)
	if err != nil {
		return discData{}, err
	}
	mbToc, err := utils.GetMusicBrainzTOC(disc)
	if err != nil {
		return discData{}, fmt.Errorf("get musicbrainz toc: %w", err)
	}
	return discData{id: id, gnuToc: gnuToc, mbToc: mbToc}, nil
}

// resolveInfo returns metadata from MusicBrainzID when set, else from a TOC lookup.
func resolveInfo(cfg *config.Config, opts Options, disc discData) (*types.DiscInfo, error) {
	if opts.MusicBrainzID != "" {
		info, err := musicbrainz.FetchReleaseByID(opts.MusicBrainzID)
		if err != nil {
			return nil, fmt.Errorf("fetch musicbrainz release %s: %w", opts.MusicBrainzID, err)
		}
		return info, nil
	}
	info, err := fetchDiscInfoConcurrently(cfg, disc.gnuToc, disc.mbToc)
	if err != nil {
		return nil, fmt.Errorf("fetch disc metadata: %w", err)
	}
	return info, nil
}

// write fetches cover art if missing, renders the sheet and persists it.
func write(cacheLocation string, info *types.DiscInfo, id, path string) (string, error) {
	if err := utils.CreateFolderIfNeeded(path); err != nil {
		return "", fmt.Errorf("create cache dir for %s: %w", path, err)
	}
	ensureCoverArt(cacheLocation, info, id)
	if err := os.WriteFile(path, renderCue(info), 0o644); err != nil {
		return "", fmt.Errorf("write cue %s: %w", path, err)
	}
	log.Printf("playlist generated at %s", path)
	return path, nil
}

// renderCue builds the CUE sheet body. Pure: no I/O, no network.
func renderCue(info *types.DiscInfo) []byte {
	var b strings.Builder
	if info.ReleaseDate != "" {
		fmt.Fprintf(&b, "REM DATE \"%s\"\n", info.ReleaseDate)
	}
	if info.Genre != "" {
		fmt.Fprintf(&b, "REM GENRE \"%s\"\n", info.Genre)
	}
	if info.CoverArtPath != "" {
		fmt.Fprintf(&b, "REM COVER \"%s\"\n", info.CoverArtPath)
	}
	fmt.Fprintf(&b, "PERFORMER \"%s\"\nTITLE \"%s\"\n", info.Artist, info.Title)
	for i, track := range info.Tracks {
		fmt.Fprintf(&b, "FILE \"cdda:///%d\" WAVE\n  TRACK %02d AUDIO\n    TITLE \"%s\"\n", i+1, i+1, track)
	}
	return []byte(b.String())
}
