package metadata

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"

	"github.com/toaster/tmpc/internal/mpd"
	"github.com/toaster/tmpc/internal/util"
)

// TODO: split into multiple fetchers and use cascade
// TODO: move caching into own CoverFetcher which might wrap a cascade
// 	     -> TODO: cache in memory (helpful when albums are split by other songs)

// CoverRepository is a repository that delivers album covers.
//
// @implements CoverFetcher
type CoverRepository struct{}

var _ CoverFetcher = (*CoverRepository)(nil)

var coverArgReplacer *strings.Replacer

func init() {
	coverArgReplacer = strings.NewReplacer("!", " ", ":", " ", "-", " ")
}

// LoadCover loads the (album) cover for a given song.
func (r *CoverRepository) LoadCover(song *mpd.Song) (fyne.Resource, error) {
	MBID := song.MBAlbumID
	if MBID == "" {
		return nil, fmt.Errorf("cannot load cover for song without MBID")
	}

	coverStream, err := r.fetchCoverFromArchive(MBID)
	if err != nil {
		log.Println("failed loading cover from coverarchive:", err)
		var artist string
		if song.Artist == song.AlbumArtist {
			artist = song.AlbumArtist
		}
		coverStream, err = r.fetchCoverFromDiscogs(artist, song.Album)
		if err != nil {
			coverStream, err = r.fetchCoverFromDiscogs(r.cleanupCoverArg(artist), r.cleanupCoverArg(song.Album))
		}
		if err != nil {
			return nil, fmt.Errorf("failed loading cover from discogs: %w", err)
		}
	}
	defer coverStream.Close()

	content, err := ioutil.ReadAll(coverStream)
	if err != nil {
		return nil, fmt.Errorf("could not read album cover %s: %w", MBID, err)
	}
	return fyne.NewStaticResource(MBID, content), nil
}

func (r *CoverRepository) fetchCoverFromArchive(MBID string) (io.ReadCloser, error) {
	return r.fetchCover(fmt.Sprintf("http://coverartarchive.org/release/%s/front", MBID))
}

type discogsAlbum struct {
	CoverImageURL string `json:"cover_image"`
}

type discogsAlbumSearchResponse struct {
	Albums []discogsAlbum `json:"results"`
}

func (r *CoverRepository) cleanupCoverArg(s string) string {
	return coverArgReplacer.Replace(s)
}

func (r *CoverRepository) fetchCoverFromDiscogs(artist, album string) (io.ReadCloser, error) {
	// TODO Discogs: support OAuth when going public -> don't use my key
	searchURL := fmt.Sprintf("https://api.discogs.com/database/search?artist=%s&release_title=%s&type=release&format=cd&key=LIPwsystksBmVhauRcJv&secret=CeSHspCISrgowJWCAWAtODZTUCKGQtRv",
		url.QueryEscape(artist), url.QueryEscape(album))
	res, err := util.HTTPGet(searchURL)
	if err != nil {
		return nil, fmt.Errorf("could not search %s - %s: %v", artist, album, err)
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
	if err := json.Unmarshal(b, &dr); err != nil {
		return nil, err
	}
	if len(dr.Albums) == 0 {
		return nil, fmt.Errorf("album not found on discogs: %s - %s", artist, album)
	}
	return r.fetchCover(dr.Albums[0].CoverImageURL)
}

func (r *CoverRepository) fetchCover(url string) (io.ReadCloser, error) {
	res, err := util.HTTPGet(url)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		defer res.Body.Close()
		b, _ := ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("could not download %s: %s", url, string(b))
	}
	return res.Body, nil
}
