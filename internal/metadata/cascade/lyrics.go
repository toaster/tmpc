package cascade

import (
	"github.com/toaster/tmpc/internal/metadata"
	"github.com/toaster/tmpc/internal/mpd"
)

// Lyrics is a metadata.LyricsFetcher cascade.
type Lyrics struct {
	fetchers []metadata.LyricsFetcher
}

var _ metadata.LyricsFetcher = (*Lyrics)(nil)

// NewLyrics returns a new Lyrics cascade.
func NewLyrics(fetchers []metadata.LyricsFetcher) *Lyrics {
	return &Lyrics{fetchers}
}

// FetchLyrics returns the lyrics of a given song.
// It tries to fetch the lyrics from the fetchers of the cascade and returns the first successful result.
//
// @implements metadata.LyricsFetcher
func (l *Lyrics) FetchLyrics(song *mpd.Song) ([]string, error) {
	var lastErr error
	for _, fetcher := range l.fetchers {
		if lyrics, err := fetcher.FetchLyrics(song); err == nil {
			return lyrics, nil
		} else {
			lastErr = err
		}
	}
	return nil, lastErr
}
