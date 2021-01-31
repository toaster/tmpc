package metadata

import (
	"crypto/sha256"
	"fmt"

	"github.com/toaster/tmpc/internal/mpd"
)

// CoverID returns a unique ID to identify the cover of a song.
func CoverID(song *mpd.Song) string {
	if song.MBAlbumID != "" {
		return song.MBAlbumID
	}
	return SongID(song)
}

// SongID returns a unique ID to identify a song.
func SongID(song *mpd.Song) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(song.File)))
}
