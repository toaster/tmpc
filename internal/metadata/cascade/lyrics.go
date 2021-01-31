package cascade

import (
	"fmt"

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
	lastErr := fmt.Errorf("could not fetch lyrics for song “%s” of “%s”", song.Title, song.Artist)
	for _, fetcher := range l.fetchers {
		lyrics, err := fetcher.FetchLyrics(song)
		if err == nil {
			return lyrics, nil
		}
		lastErr = err
	}
	return nil, lastErr
}
