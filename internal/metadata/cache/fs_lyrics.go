package cache

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/toaster/tmpc/internal/metadata"
	"github.com/toaster/tmpc/internal/mpd"
)

// FSLyrics is a file system cache for metadata.LyricsFetcher.
//
// @implements metadata.LyricsFetcher
type FSLyrics struct {
	dir     string
	fetcher metadata.LyricsFetcher
}

var _ metadata.LyricsFetcher = (*FSLyrics)(nil)

// NewFSLyrics returns a new FSLyrics cover cache.
func NewFSLyrics(fetcher metadata.LyricsFetcher) *FSLyrics {
	dir, err := TmpDir()
	if err != nil {
		log.Fatal("cannot access temp dir:", err)
	}
	return &FSLyrics{dir, fetcher}
}

// FetchLyrics tries to fetch the lyrics from the FSLyrics cache and uses the fetcher if it fails.
//
// @implements metadata.LyricsFetcher
func (f *FSLyrics) FetchLyrics(song *mpd.Song) ([]string, error) {
	path := filepath.Join(f.dir, metadata.SongID(song))

	content, err := ioutil.ReadFile(path)
	if err == nil {
		return strings.Split(string(content), "\n"), nil
	}

	lyrics, err := f.fetcher.FetchLyrics(song)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(path, []byte(strings.Join(lyrics, "\n")), 0600)
	if err != nil {
		log.Printf("could not write %s: %v", path, err)
	}
	return lyrics, nil
}
