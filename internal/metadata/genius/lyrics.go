package genius

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/toaster/tmpc/internal/metadata"
	"github.com/toaster/tmpc/internal/mpd"
	"github.com/toaster/tmpc/internal/util"
)

var _ metadata.LyricsFetcher = (*Lyrics)(nil)

// Lyrics is a LyricsFetcher that delivers the lyrics of a song.
//
// @implements metadata.LyricsFetcher
type Lyrics struct {
	accessToken            string
	albumIDsByMBAlbumIDs   map[string]int
	artistIDsByMBArtistIDs map[string]int
	trackIDsByMBTrackIDs   map[string]int
}

// NewLyrics creates a new Genius lyrics repository using the given access token.
func NewLyrics(accessToken string) *Lyrics {
	return &Lyrics{
		accessToken:            accessToken,
		albumIDsByMBAlbumIDs:   map[string]int{},
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
	lyrics, err := l.fetchLyrics(info)
	if err != nil {
		return nil, err
	}
	return lyrics, nil
}

func (l *Lyrics) fetchLyrics(info *song) ([]string, error) {
	if info.LyricsURL == "" {
		return nil, fmt.Errorf("no lyrics provided for song “%s” of “%s” (%d)", info.Title, info.PrimaryArtist.Name, info.ID)
	}

	doc, err := util.HTTPGetHTML(info.LyricsURL)
	if err != nil {
		return nil, err
	}
	lyrics := l.findLyricsInHTML(doc)
	if lyrics == nil {
		return nil, fmt.Errorf("could not find lyrics for “%s” in %s", info.Title, info.LyricsURL)
	}
	return metadata.ExtractLyricsFromHTML(lyrics, nil), nil
}

func (l *Lyrics) findLyricsInHTML(n *html.Node) []*html.Node {
	if n.Type == html.ElementNode && n.DataAtom == atom.Div {
		if metadata.NodeParamsMatch(n, map[string][]metadata.Matcher{
			"class":                 {metadata.NewExactMatcher("lyrics"), metadata.NewPrefixMatcher("Lyrics__Container")},
			"data-lyrics-container": {metadata.NewExactMatcher("true")},
		}) {
			return []*html.Node{n}
		}
	}

	nodes := ([]*html.Node)(nil)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		additionalNodes := l.findLyricsInHTML(c)
		if additionalNodes != nil {
			nodes = append(nodes, additionalNodes...)
		}
	}
	return nodes
}

func (l *Lyrics) gatherSongInfo(song *mpd.Song) (info *song, err error) {
	trackID := l.trackIDsByMBTrackIDs[song.MBTrackID]
	if trackID == 0 {
		artistID := l.artistIDsByMBArtistIDs[song.MBArtistID]
		if artistID == 0 {
			artistID, err = l.searchArtist(song.Artist)
			if err != nil {
				return
			}
			l.artistIDsByMBArtistIDs[song.MBArtistID] = artistID
		}
		trackID, err = l.gatherTrackID(artistID, song.Title)
		if err != nil {
			return
		}
	}

	result := songResult{}
	err = util.HTTPGetJSON(
		l.uri(fmt.Sprintf("songs/%d", trackID), ""),
		&result,
	)
	if err != nil {
		return
	}
	info = &result.Response.Song
	return
}

func (l *Lyrics) gatherTrackID(artistID int, title string) (int, error) {
	page := 1
	for {
		result := songsResult{}
		err := util.HTTPGetJSON(
			l.uri(fmt.Sprintf("artists/%d/songs", artistID), fmt.Sprintf("per_page=50&page=%d", page)),
			&result,
		)
		if err != nil {
			return 0, err
		}
		if result.Meta.Status != http.StatusOK {
			return 0, fmt.Errorf("searching for s failed: %d - %s", result.Meta.Status, result.Meta.Message)
		}

		for _, s := range result.Response.Songs {
			if strings.ToLower(s.Title) == strings.ToLower(title) {
				return s.ID, nil
			}
		}
		page = result.Response.NextPage
		if page == 0 {
			return 0, fmt.Errorf("could not find song “%s” for artist ID “%d”", title, artistID)
		}
	}
}

func (l *Lyrics) searchArtist(artist string) (int, error) {
	result := searchResult{}
	err := util.HTTPGetJSON(
		l.uri("search", fmt.Sprintf("q=%s", url.QueryEscape(`"`+artist+`"`))),
		&result,
	)
	if err != nil {
		return 0, err
	}
	if result.Meta.Status != http.StatusOK {
		return 0, fmt.Errorf("searching for song failed: %d - %s", result.Meta.Status, result.Meta.Message)
	}

	for _, hit := range result.Response.Hits {
		if strings.ToLower(hit.Result.PrimaryArtist.Name) == strings.ToLower(artist) {
			return hit.Result.PrimaryArtist.ID, nil
		}
	}

	return 0, fmt.Errorf("could not find “%s” in: %v", artist, result.Response.Hits)
}

func (l *Lyrics) uri(path, query string) string {
	return fmt.Sprintf("https://api.genius.com/%s?%s&access_token=%s", path, query, l.accessToken)
}

type apiResult struct {
	Meta apiResultMeta `json:"meta"`
}

type apiResultMeta struct {
	Message string `json:"message,omitempty"`
	Status  int    `json:"status"`
}

type searchHit struct {
	Result song `json:"result"`
}

type searchHitArtist struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type searchResponse struct {
	Hits []searchHit `json:"hits"`
}

type searchResult struct {
	apiResult
	Response searchResponse `json:"response"`
}

type song struct {
	ID            int             `json:"id"`
	LyricsURL     string          `json:"url"`
	PrimaryArtist searchHitArtist `json:"primary_artist"`
	Title         string          `json:"title"`
}

type songResponse struct {
	Song song `json:"song"`
}

type songResult struct {
	apiResult
	Response songResponse `json:"response"`
}

type songsResponse struct {
	NextPage int    `json:"next_page,omitempty"`
	Songs    []song `json:"songs"`
}

type songsResult struct {
	apiResult
	Response songsResponse `json:"response"`
}
