package archiveorg

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"fyne.io/fyne/v2"

	"github.com/toaster/tmpc/internal/metadata"
	"github.com/toaster/tmpc/internal/mpd"
	"github.com/toaster/tmpc/internal/util"
)

// Cover is a CoverFetcher that delivers the cover of a song.
//
// @implements metadata.CoverFetcher
type Cover struct{}

var _ metadata.CoverFetcher = (*Cover)(nil)

// NewCover returns a new Cover.
func NewCover() *Cover {
	return &Cover{}
}

// LoadCover loads the cover of a song.
//
// @implements metadata.CoverFetcher
func (c *Cover) LoadCover(song *mpd.Song) (fyne.Resource, error) {
	MBID := song.MBAlbumID
	if MBID == "" {
		return nil, fmt.Errorf("cannot load cover for song without MBID")
	}

	url := fmt.Sprintf("http://coverartarchive.org/release/%s/front", MBID)
	res, err := util.HTTPGet(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("could not download %s: %s", url, string(b))
	}

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read album cover %s: %w", MBID, err)
	}
	return fyne.NewStaticResource(MBID, content), nil
}
