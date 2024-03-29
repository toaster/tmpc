package cache

import (
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"

	"github.com/toaster/tmpc/internal/metadata"
	"github.com/toaster/tmpc/internal/mpd"
	"github.com/toaster/tmpc/internal/util"
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

	content, err := os.ReadFile(imgPath)
	if err == nil {
		return fyne.NewStaticResource(id, content), nil
	}

	cover, err := f.fetcher.LoadCover(song)
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(imgPath, cover.Content(), util.PermUserRead|util.PermUserWrite)
	if err != nil {
		log.Printf("could not write %s: %v", imgPath, err)
	}
	return cover, nil
}
