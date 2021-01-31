package metadata

import (
	"fyne.io/fyne/v2"

	"github.com/toaster/tmpc/internal/mpd"
)

// CoverFetcher is a repository that delivers cover art for a song.
type CoverFetcher interface {
	LoadCover(song *mpd.Song) (fyne.Resource, error)
}
