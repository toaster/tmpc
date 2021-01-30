package metadata

import "github.com/toaster/tmpc/internal/mpd"

// LyricsFetcher is a repository that delivers lyrics for a song.
type LyricsFetcher interface {
	FetchLyrics(*mpd.Song) ([]string, error)
}
