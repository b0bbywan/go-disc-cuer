package cue

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/b0bbywan/go-disc-cuer/config"
	"github.com/b0bbywan/go-disc-cuer/gnudb"
	"github.com/b0bbywan/go-disc-cuer/musicbrainz"
	"github.com/b0bbywan/go-disc-cuer/types"
	"github.com/b0bbywan/go-disc-cuer/utils"
)

// coverArtURL is the Cover Art Archive base. It is a var rather than a const so
// tests can redirect it to a local stub server.
var coverArtURL = "https://coverartarchive.org/release"

// ensureCoverArt downloads the front cover into the cache when info has none.
// Failures are logged, not fatal: a sheet without cover is still valid.
func ensureCoverArt(cacheLocation string, info *types.DiscInfo, id string) {
	if info.CoverArtPath != "" {
		return
	}
	dst := utils.CacheCoverArtPath(cacheLocation, id)
	if err := fetchCoverArt(info.ID, dst); err != nil {
		log.Printf("cover art: %v", err)
		return
	}
	info.CoverArtPath = dst
}

// fetchCoverArt saves the Cover Art Archive front image for a MusicBrainz id.
func fetchCoverArt(mbID, dst string) error {
	resp, err := http.Get(fmt.Sprintf("%s/%s/front", coverArtURL, mbID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}

	file, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// fetchDiscInfoConcurrently queries GNUDB and MusicBrainz in parallel by TOC.
func fetchDiscInfoConcurrently(cfg *config.Config, gnuToc, mbToc string) (*types.DiscInfo, error) {
	var wg sync.WaitGroup
	var gnudbInfo, mbInfo *types.DiscInfo
	var gnudbErr, mbErr error

	wg.Add(2)
	go func() {
		defer wg.Done()
		gnudbInfo, gnudbErr = gnudb.FetchDiscInfo(cfg, strings.ReplaceAll(gnuToc, " ", "+"))
	}()
	go func() {
		defer wg.Done()
		mbInfo, mbErr = musicbrainz.FetchReleaseByToc(strings.ReplaceAll(mbToc, " ", "+"))
	}()
	wg.Wait()

	return selectDiscInfo(gnudbInfo, gnudbErr, mbInfo, mbErr)
}

// selectDiscInfo prefers the GNUDB body but always keeps the MusicBrainz id,
// which the cover art lookup needs. It fails only when both sources fail.
func selectDiscInfo(gnudbInfo *types.DiscInfo, gnudbErr error, mbInfo *types.DiscInfo, mbErr error) (*types.DiscInfo, error) {
	if gnudbErr != nil && mbErr != nil {
		return nil, fmt.Errorf("gnudb: %w; musicbrainz: %w", gnudbErr, mbErr)
	}

	info := &types.DiscInfo{}
	switch {
	case gnudbErr == nil:
		*info = *gnudbInfo
	case mbErr == nil:
		*info = *mbInfo
	}
	if mbErr == nil {
		info.ID = mbInfo.ID
	}
	return info, nil
}
