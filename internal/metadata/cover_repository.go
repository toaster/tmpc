package metadata

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"

	"github.com/toaster/tmpc/internal/mpd"
	"github.com/toaster/tmpc/internal/util"
)

// CoverRepository is a repository that delivers album covers.
type CoverRepository struct{}

var coverArgReplacer *strings.Replacer

func init() {
	coverArgReplacer = strings.NewReplacer("!", " ", ":", " ", "-", " ")
}

// LoadCover loads the (album) cover for a given song and runs the callback with the image.
// The callback might be run twice: once with a default and later with the real cover.
// This way the cover can be displayed immediately and updated once the real data is available.
func (r *CoverRepository) LoadCover(song *mpd.Song, coverDefault fyne.Resource, callback func(fyne.Resource)) {
	callback(coverDefault)
	if song == nil {
		return
	}
	go func() {
		tmpDir := filepath.Join(os.TempDir(), "tmpc")
		if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
			log.Println("could not create tmp dir:", err)
			return
		}

		MBID := song.MBAlbumID
		// TODO cache in memory (helpful when albums are splitted by other songs)
		if MBID != "" {
			imgPath := filepath.Join(tmpDir, MBID)
			imgFile, err := os.Open(imgPath)
			var imgReader io.Reader
			if err == nil {
				defer imgFile.Close()
				imgReader = imgFile
			} else {
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
						log.Println("failed loading cover from discogs:", err)
						return
					}
					// TODO try other services
				}

				defer coverStream.Close()
				imgReader = coverStream

				imgFile, err = os.Create(imgPath)
				if err != nil {
					log.Printf("could not create %s: %v", imgPath, err)
				} else {
					defer imgFile.Close()
					imgReader = io.TeeReader(imgReader, imgFile)
				}
			}

			content, err := ioutil.ReadAll(imgReader)
			if err != nil {
				log.Printf("could not read album cover %s: %v", MBID, err)
				return
			}
			callback(fyne.NewStaticResource(MBID, content))
		}
	}()
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
