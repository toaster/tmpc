package metadata

import (
	"crypto/sha256"
	"fmt"

	"fyne.io/fyne/v2"

	"github.com/toaster/tmpc/internal/mpd"
)

// CoverFetcher is a repository that delivers cover art for a song.
type CoverFetcher interface {
	LoadCover(song *mpd.Song) (fyne.Resource, error)
}

// CoverID returns a unique ID to identify the cover of a song.
func CoverID(song *mpd.Song) string {
	id := song.MBAlbumID
	if id == "" {
		id = fmt.Sprintf("%x", sha256.Sum256([]byte(song.File)))
	}
	return id
}
