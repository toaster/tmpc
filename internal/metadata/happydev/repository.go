package happydev

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/toaster/tmpc/internal/metadata"
	"github.com/toaster/tmpc/internal/mpd"
	"github.com/toaster/tmpc/internal/util"
)

// Repository is a LyricsFetcher that delivers the lyrics of a song.
//
// @implements metadata.LyricsFetcher
type Repository struct {
	apiKey string
}

var _ metadata.LyricsFetcher = (*Repository)(nil)

// NewRepository creates a new happy.dev repository using the given API key.
func NewRepository(apiKey string) *Repository {
	return &Repository{apiKey}
}

// FetchLyrics fetches the lyrics for a song.
//
// @implements metadata.LyricsFetcher
func (r *Repository) FetchLyrics(song *mpd.Song) ([]string, error) {
	if song == nil {
		return []string{}, nil
	}

	info, err := r.findSong(song.Artist, song.Album, song.Title)
	if err != nil {
		return nil, err
	}
	lyrics, err := r.fetchLyrics(info)
	if err != nil {
		return nil, err
	}
	return lyrics, nil
}

func (r *Repository) fetchLyrics(info *songInfo) ([]string, error) {
	if !info.HasLyrics {
		// TODO: error? special value?
		return nil, nil
	}

	result := lyricsResult{}
	err := util.HTTPGetJSON(
		r.uri(fmt.Sprintf("/artists/%d/albums/%d/tracks/%d/lyrics", info.ArtistID, info.AlbumID, info.TrackID), ""),
		&result,
	)
	if err != nil {
		return nil, err
	}
	if result.Length == 0 {
		// TODO: error? special value?
		return nil, nil
	}
	// TODO: handle multiple results

	return strings.Split(result.Entries[0].Lyrics, "\n"), nil
}

func (r *Repository) findSong(artist, album, title string) (*songInfo, error) {
	result := searchResult{}
	err := util.HTTPGetJSON(r.uri("", fmt.Sprintf("q=%s%%20%s", url.QueryEscape(artist), url.QueryEscape(title))), &result)
	if err != nil {
		return nil, err
	}
	if result.Length == 0 {
		return nil, fmt.Errorf("could not find song “%s” of “%s”", title, artist)
	}
	// TODO: bad case
	// result.Success == false
	for _, hit := range result.Hits {
		if hit.Album == album {
			return &hit, nil
		}
	}
	// TODO: check song name and artist name for perfect match
	// TODO: log imperfect match
	return &result.Hits[0], nil
}

func (r *Repository) uri(path, query string) string {
	return fmt.Sprintf("https://api.happi.dev/v1/music%s?%s&apikey=%s", path, query, r.apiKey)
}

type lyricsInfo struct {
	Lyrics string `json:"lyrics"`
}

type lyricsResult struct {
	apiResult
	Entries []lyricsInfo `json:"result"`
}

type apiResult struct {
	Length  int  `json:"length"`
	Success bool `json:"success"`
}

type searchResult struct {
	apiResult
	Hits []songInfo `json:"result"`
}

type songInfo struct {
	Album     string `json:"album"`
	AlbumID   int    `json:"id_album"`
	Artist    string `json:"artist"`
	ArtistID  int    `json:"id_artist"`
	HasLyrics bool   `json:"haslyrics"`
	Track     string `json:"track"`
	TrackID   int    `json:"id_track"`
}
