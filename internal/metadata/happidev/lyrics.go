package happidev

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/toaster/tmpc/internal/metadata"
	"github.com/toaster/tmpc/internal/mpd"
	"github.com/toaster/tmpc/internal/util"
)

// Lyrics is a LyricsFetcher that delivers the lyrics of a song.
//
// @implements metadata.LyricsFetcher
type Lyrics struct {
	albumIDsByMBAlbumIDs   map[string]int
	apiKey                 string
	artistIDsByMBArtistIDs map[string]int
	trackIDsByMBTrackIDs   map[string]int
}

var _ metadata.LyricsFetcher = (*Lyrics)(nil)

// NewLyrics creates a new happi.dev lyrics repository using the given API key.
func NewLyrics(apiKey string) *Lyrics {
	return &Lyrics{
		albumIDsByMBAlbumIDs:   map[string]int{},
		apiKey:                 apiKey,
		artistIDsByMBArtistIDs: map[string]int{},
		trackIDsByMBTrackIDs:   map[string]int{},
	}
}

// FetchLyrics fetches the lyrics for a song.
//
// @implements metadata.LyricsFetcher
func (l *Lyrics) FetchLyrics(song *mpd.Song) ([]string, error) {
	if song == nil {
		return []string{}, nil
	}

	info, err := l.gatherSongInfo(song)
	if err != nil {
		return nil, err
	}
	lyrics, err := l.fetchLyricsForSongInfo(info)
	if err != nil {
		return nil, err
	}
	return lyrics, nil
}

func (l *Lyrics) fetchLyricsForSongInfo(info *songInfo) ([]string, error) {
	if !info.HasLyrics {
		return nil, fmt.Errorf("no lyrics provided for song “%s” of “%s”", info.Track, info.Artist)
	}

	result := lyricsResult{}
	err := util.HTTPGetJSON(
		l.uri(fmt.Sprintf("/artists/%d/albums/%d/tracks/%d/lyrics", info.ArtistID, info.AlbumID, info.TrackID), ""),
		&result,
	)
	if err != nil {
		return nil, err
	}
	if result.Length == 0 {
		return nil, fmt.Errorf("fetched lyrics result was empty for song “%s” of “%s”", info.Track, info.Artist)
	}

	return strings.Split(result.Record.Lyrics, "\n"), nil
}

func (l *Lyrics) fetchSongInfo(artistID, albumID, trackID int) (*songInfo, error) {
	result := trackResult{}
	err := util.HTTPGetJSON(
		l.uri(fmt.Sprintf("/artists/%d/albums/%d/tracks/%d", artistID, albumID, trackID), ""),
		&result,
	)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("failed to fetch track: %s", result.Error)
	}
	if result.Length == 0 {
		return nil, fmt.Errorf("track has no info")
	}

	return &songInfo{
		Album:     result.Info.Album,
		AlbumID:   albumID,
		Artist:    result.Info.Artist,
		ArtistID:  artistID,
		HasLyrics: result.Info.HasLyrics,
		Track:     result.Info.Track,
		TrackID:   trackID,
	}, nil
}

func (l *Lyrics) gatherAlbumID(artistID int, album string) (int, error) {
	result := albumsResult{}
	err := util.HTTPGetJSON(
		l.uri(fmt.Sprintf("/artists/%d/albums", artistID), ""),
		&result,
	)
	if err != nil {
		return 0, err
	}
	if !result.Success {
		return 0, fmt.Errorf("failed to fetch albums: %s", result.Error)
	}
	if result.Length == 0 || result.Info.Length == 0 {
		return 0, fmt.Errorf("artist has no albums")
	}

	lowerAlbum := strings.ToLower(album)
	for _, hit := range result.Info.Albums {
		if strings.ToLower(hit.Album) == lowerAlbum {
			return hit.AlbumID, nil
		}
	}
	for _, hit := range result.Info.Albums {
		lowerHit := strings.ToLower(hit.Album)
		if strings.Contains(lowerHit, lowerAlbum) || strings.Contains(lowerAlbum, lowerHit) {
			return hit.AlbumID, nil
		}
	}
	for _, lang := range []string{"en", "de"} {
		reducedAlbum := metadata.ReducedTitle(lowerAlbum, lang)
		for _, hit := range result.Info.Albums {
			reducedHit := metadata.ReducedTitle(strings.ToLower(hit.Album), lang)
			if strings.Contains(reducedHit, reducedAlbum) || strings.Contains(reducedAlbum, reducedHit) {
				return hit.AlbumID, nil
			}
		}
	}

	return 0, fmt.Errorf("could not find album “%s” in %v", album, result.Info)
}

func (l *Lyrics) gatherSongInfo(song *mpd.Song) (*songInfo, error) {
	album := song.Album
	artist := song.Artist
	title := song.Title
	info, err := l.searchTrack(artist, title, album)

	if err != nil {
		artistID := l.artistIDsByMBArtistIDs[song.MBArtistID]
		if artistID == 0 {
			artistID, err = l.searchArtist(artist, album)
			if err != nil {
				return nil, err
			}
		}

		albumID := l.albumIDsByMBAlbumIDs[song.MBAlbumID]
		if albumID == 0 {
			albumID, err = l.gatherAlbumID(artistID, album)
			if err != nil {
				return nil, err
			}
		}

		trackID := l.trackIDsByMBTrackIDs[song.MBTrackID]
		if trackID == 0 {
			trackID, err = l.gatherTrackID(artistID, albumID, title)
			if err != nil {
				return nil, err
			}
		}

		info, err = l.fetchSongInfo(artistID, albumID, trackID)
		if err != nil {
			return nil, err
		}
	}

	l.albumIDsByMBAlbumIDs[song.MBAlbumID] = info.AlbumID
	l.artistIDsByMBArtistIDs[song.MBArtistID] = info.ArtistID
	l.trackIDsByMBTrackIDs[song.MBTrackID] = info.TrackID
	return info, nil
}

func (l *Lyrics) gatherTrackID(artistID, albumID int, track string) (int, error) {
	result := tracksResult{}
	err := util.HTTPGetJSON(
		l.uri(fmt.Sprintf("/artists/%d/albums/%d/tracks", artistID, albumID), ""),
		&result,
	)
	if err != nil {
		return 0, err
	}
	if !result.Success {
		return 0, fmt.Errorf("failed to fetch tracks: %s", result.Error)
	}
	if result.Length == 0 || result.Info.Length == 0 {
		return 0, fmt.Errorf("album has no tracks")
	}

	lowerTrack := strings.ToLower(track)
	for _, hit := range result.Info.Tracks {
		if strings.ToLower(hit.Track) == lowerTrack {
			return hit.TrackID, nil
		}
	}
	for _, lang := range []string{"en", "de"} {
		reducedTrack := metadata.ReducedTitle(lowerTrack, lang)
		for _, hit := range result.Info.Tracks {
			reducedHit := metadata.ReducedTitle(strings.ToLower(hit.Track), lang)
			if strings.Contains(reducedHit, reducedTrack) || strings.Contains(reducedTrack, reducedHit) {
				return hit.TrackID, nil
			}
		}
	}

	return 0, fmt.Errorf("could not find track “%s” in %v", track, result.Info)
}

func (l *Lyrics) searchArtist(artist string, album string) (int, error) {
	result := artistSearchResult{}
	err := util.HTTPGetJSON(
		l.uri("", fmt.Sprintf("type=artist&limit=50&q=%s", url.QueryEscape(artist))),
		&result,
	)
	if err != nil {
		return 0, err
	}
	if !result.Success {
		return 0, fmt.Errorf("searching for artist failed: %s", result.Error)
	}
	if result.Length == 0 {
		return 0, fmt.Errorf("could not find artist “%s”", artist)
	}

	for _, hit := range result.Hits {
		if strings.EqualFold(hit.Artist, artist) {
			albumID, _ := l.gatherAlbumID(hit.ArtistID, album)
			if albumID != 0 {
				return hit.ArtistID, nil
			}
		}
	}

	return 0, fmt.Errorf("could not find “%s” (with album “%s”): %v", artist, album, result.Hits)
}

func (l *Lyrics) searchTrack(artist string, title string, album string) (*songInfo, error) {
	result := trackSearchResult{}
	err := util.HTTPGetJSON(
		l.uri(
			"",
			fmt.Sprintf("type=track&limit=50&q=%s%%20%s", url.QueryEscape(artist), url.QueryEscape(title)),
		),
		&result,
	)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("searching for song failed: %s", result.Error)
	}
	if result.Length == 0 {
		return nil, fmt.Errorf("could not find song “%s” of “%s”", title, artist)
	}

	for _, hit := range result.Hits {
		if strings.EqualFold(hit.Album, album) &&
			strings.EqualFold(hit.Artist, artist) &&
			strings.EqualFold(hit.Track, title) {
			return &hit, nil
		}
	}

	return nil, fmt.Errorf("could not find “%s” of “%s” in: %v", title, artist, result.Hits)
}

func (l *Lyrics) uri(path, query string) string {
	return fmt.Sprintf("https://api.happi.dev/v1/music%s?%s&apikey=%s", path, query, l.apiKey)
}

type albumsResult struct {
	apiResult
	Info artistAlbumsInfo `json:"result"`
}

type albumTrackInfo struct {
	Track   string `json:"track"`
	TrackID int    `json:"id_track"`
}

type albumTracksInfo struct {
	Tracks []albumTrackInfo `json:"tracks"`
	Length int              `json:"length"`
}

type apiResult struct {
	Error   string `json:"error"`
	Length  int    `json:"length"`
	Success bool   `json:"success"`
}

type artistAlbumInfo struct {
	Album   string `json:"album"`
	AlbumID int    `json:"id_album"`
}

type artistAlbumsInfo struct {
	Albums []artistAlbumInfo `json:"albums"`
	Length int               `json:"length"`
}

type artistInfo struct {
	Artist   string `json:"artist"`
	ArtistID int    `json:"id_artist"`
}

type artistSearchResult struct {
	apiResult
	Hits []artistInfo `json:"result"`
}

type lyricsInfo struct {
	Lyrics string `json:"lyrics"`
}

type lyricsResult struct {
	apiResult
	Record lyricsInfo `json:"result"`
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

type trackInfo struct {
	Album     string `json:"album"`
	Artist    string `json:"artist"`
	HasLyrics bool   `json:"haslyrics"`
	Track     string `json:"track"`
	TrackID   int    `json:"id_track"`
}

type trackResult struct {
	apiResult
	Info trackInfo `json:"result"`
}

type trackSearchResult struct {
	apiResult
	Hits []songInfo `json:"result"`
}

type tracksResult struct {
	apiResult
	Info albumTracksInfo `json:"result"`
}
