package cascade

import (
	"fmt"

	"fyne.io/fyne/v2"

	"github.com/toaster/tmpc/internal/metadata"
	"github.com/toaster/tmpc/internal/mpd"
)

// Cover is a metadata.CoverFetcher cascade.
type Cover struct {
	fetchers []metadata.CoverFetcher
}

var _ metadata.CoverFetcher = (*Cover)(nil)

// NewCover returns a new Cover cascade.
func NewCover(fetchers []metadata.CoverFetcher) *Cover {
	return &Cover{fetchers}
}

// LoadCover returns the cover of a given song.
// It tries to fetch the cover from the fetchers of the cascade and returns the first successful result.
//
// @implements metadata.LyricsFetcher
func (c *Cover) LoadCover(song *mpd.Song) (fyne.Resource, error) {
	lastErr := fmt.Errorf("could not load cover for song “%s” of “%s”", song.Title, song.Artist)
	for _, fetcher := range c.fetchers {
		cover, err := fetcher.LoadCover(song)
		if err == nil {
			return cover, nil
		}
		lastErr = err
	}
	return nil, lastErr
}
