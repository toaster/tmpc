package discogs

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"

	"github.com/toaster/tmpc/internal/metadata"
	"github.com/toaster/tmpc/internal/mpd"
	"github.com/toaster/tmpc/internal/util"
)

var coverArgReplacer *strings.Replacer

func init() {
	coverArgReplacer = strings.NewReplacer("!", " ", ":", " ", "-", " ")
}

// Cover is a CoverFetcher that delivers the cover of a song.
//
// @implements metadata.CoverFetcher
type Cover struct {
	key    string
	secret string
}

var _ metadata.CoverFetcher = (*Cover)(nil)

// NewCover returns a new Cover.
func NewCover(key, secret string) *Cover {
	return &Cover{key, secret}
}

// LoadCover loads the cover of a song.
//
// @implements metadata.CoverFetcher
func (c *Cover) LoadCover(song *mpd.Song) (fyne.Resource, error) {
	var artist string
	if song.Artist == song.AlbumArtist {
		artist = song.AlbumArtist
	}
	coverStream, err := c.fetchCoverFromDiscogs(artist, song.Album)
	if err != nil {
		coverStream, err = c.fetchCoverFromDiscogs(c.cleanupCoverArg(artist), c.cleanupCoverArg(song.Album))
	}
	if err != nil {
		return nil, fmt.Errorf("failed loading cover: %w", err)
	}
	defer coverStream.Close()

	content, err := ioutil.ReadAll(coverStream)
	if err != nil {
		return nil, fmt.Errorf("could not read cover for song “%s” of “%s”: %w", song.Title, song.Artist, err)
	}
	return fyne.NewStaticResource(metadata.CoverID(song), content), nil
}

func (c *Cover) cleanupCoverArg(s string) string {
	return coverArgReplacer.Replace(s)
}

func (c *Cover) fetchCoverFromDiscogs(artist, album string) (io.ReadCloser, error) {
	searchURL := fmt.Sprintf("https://api.discogs.com/database/search?artist=%s&release_title=%s&type=release&format=cd&key=%s&secret=%s",
		url.QueryEscape(artist), url.QueryEscape(album), url.QueryEscape(c.key), url.QueryEscape(c.secret))
	res, err := util.HTTPGet(searchURL)
	if err != nil {
		return nil, fmt.Errorf("could not search %s - %s: %w", artist, album, err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("could not search %s - %s: %s", artist, album, string(b))
	}

	var dr discogsAlbumSearchResponse
	if err = json.Unmarshal(b, &dr); err != nil {
		return nil, err
	}
	if len(dr.Albums) == 0 {
		return nil, fmt.Errorf("album not found on discogs: %s - %s", artist, album)
	}

	coverURL := dr.Albums[0].CoverImageURL
	res, err = util.HTTPGet(coverURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ = ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("could not download %s: %s", coverURL, string(b))
	}

	return res.Body, nil
}

type discogsAlbum struct {
	CoverImageURL string `json:"cover_image"`
}

type discogsAlbumSearchResponse struct {
	Albums []discogsAlbum `json:"results"`
}
