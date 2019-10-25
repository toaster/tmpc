package repo

import (
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/toaster/tmpc/internal/mpd"
	"github.com/toaster/tmpc/internal/util"
)

// CoverRepository is a repository that delivers album covers.
type CoverRepository struct{}

var coverDefault draw.Image
var coverArgReplacer *strings.Replacer

func init() {
	coverDefault = image.NewGray16(image.Rect(0, 0, 10, 10))
	draw.Draw(coverDefault, image.Rect(0, 0, 10, 10), image.White, image.Pt(0, 0), draw.Over)
	coverArgReplacer = strings.NewReplacer("!", " ", ":", " ", "-", " ")
}

// LoadCover returns the (album) cover for a given song.
func (r *CoverRepository) LoadCover(song *mpd.Song) image.Image {
	if song == nil {
		return coverDefault
	}
	tmpDir := fmt.Sprintf("%s/tmpc", os.TempDir())
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		log.Println("could not create tmp dir:", err)
		return coverDefault
	}

	MBID := song.MBAlbumID
	// TODO cache in memory (helpful when albums are splitted by other songs)
	var cover image.Image
	if MBID != "" {
		imgPath := fmt.Sprintf("%s/%s", tmpDir, MBID)
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
					return coverDefault
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

		cover, _, err = image.Decode(imgReader)
		if err != nil {
			log.Printf("could not decode %s: %v", MBID, err)
		}
	}
	if cover == nil {
		cover = coverDefault
	}
	return cover
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
