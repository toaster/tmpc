package cache

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"fyne.io/fyne/v2"

	"github.com/toaster/tmpc/internal/metadata"
	"github.com/toaster/tmpc/internal/mpd"
)

// FSCover is a file system cache for metadata.CoverFetcher.
//
// @implements metadata.CoverFetcher
type FSCover struct {
	fetcher metadata.CoverFetcher
}

var _ metadata.CoverFetcher = (*FSCover)(nil)

// NewFSCover returns a new FSCover cover cache.
func NewFSCover(fetcher metadata.CoverFetcher) *FSCover {
	return &FSCover{fetcher}
}

// LoadCover tries to load the cover from the FSCover cache and uses the fetcher if it fails.
//
// @implements metadata.CoverFetcher
func (f *FSCover) LoadCover(song *mpd.Song) (fyne.Resource, error) {
	dir, err := TmpDir()
	if err != nil {
		return nil, err
	}

	id := metadata.CoverID(song)
	imgPath := filepath.Join(dir, id)

	content, err := ioutil.ReadFile(imgPath)
	if err == nil {
		return fyne.NewStaticResource(id, content), nil
	}

	cover, err := f.fetcher.LoadCover(song)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(imgPath, cover.Content(), 0600)
	if err != nil {
		log.Printf("could not write %s: %v", imgPath, err)
	}
	return cover, nil
}
