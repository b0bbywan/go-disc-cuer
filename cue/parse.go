package cue

import (
	"fmt"
	"os"
	"strings"

	"github.com/b0bbywan/go-disc-cuer/types"
)

// Read returns the metadata of the CUE sheet at path, as written by Generate.
// The MusicBrainz id is not kept in the sheet, so Info.ID is empty.
func Read(path string) (*types.DiscInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read cue %s: %w", path, err)
	}
	return parseCue(data), nil
}

// parseCue reads back what renderCue writes. Unknown lines are skipped. Pure:
// no I/O.
func parseCue(data []byte) *types.DiscInfo {
	info := &types.DiscInfo{}
	inTrack := false
	for _, line := range strings.Split(string(data), "\n") {
		keyword, value, _ := strings.Cut(strings.TrimSpace(line), " ")
		switch keyword {
		case "REM":
			field, value, _ := strings.Cut(value, " ")
			switch field {
			case "DATE":
				info.ReleaseDate = unquote(value)
			case "GENRE":
				info.Genre = unquote(value)
			case "COVER":
				info.CoverArtPath = unquote(value)
			}
		case "PERFORMER":
			if !inTrack {
				info.Artist = unquote(value)
			}
		case "TRACK":
			inTrack = true
			info.Tracks = append(info.Tracks, "")
		case "TITLE":
			if inTrack {
				info.Tracks[len(info.Tracks)-1] = unquote(value)
			} else {
				info.Title = unquote(value)
			}
		}
	}
	return info
}

// unquote strips the quotes around a CUE value. renderCue does not escape
// quotes, so the value runs from the first quote to the last one.
func unquote(s string) string {
	s = strings.TrimSpace(s)
	first, last := strings.Index(s, `"`), strings.LastIndex(s, `"`)
	if first == -1 || first == last {
		return s
	}
	return s[first+1 : last]
}
