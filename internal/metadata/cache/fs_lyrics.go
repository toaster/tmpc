package cache

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/toaster/tmpc/internal/metadata"
	"github.com/toaster/tmpc/internal/mpd"
	"github.com/toaster/tmpc/internal/util"
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
		panic(fmt.Errorf("cannot access temp dir: %w", err))
	}
	return &FSLyrics{dir, fetcher}
}

// FetchLyrics tries to fetch the lyrics from the FSLyrics cache and uses the fetcher if it fails.
//
// @implements metadata.LyricsFetcher
func (f *FSLyrics) FetchLyrics(song *mpd.Song) ([]string, error) {
	path := filepath.Join(f.dir, metadata.SongID(song))

	content, err := os.ReadFile(path)
	if err == nil {
		return strings.Split(string(content), "\n"), nil
	}

	lyrics, err := f.fetcher.FetchLyrics(song)
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(path, []byte(strings.Join(lyrics, "\n")), util.PermUserRead|util.PermUserWrite)
	if err != nil {
		log.Printf("could not write %s: %v", path, err)
	}
	return lyrics, nil
}
