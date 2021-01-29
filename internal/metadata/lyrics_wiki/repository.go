package lyrics_wiki

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"

	"github.com/toaster/tmpc/internal/metadata"
	"github.com/toaster/tmpc/internal/mpd"
	"github.com/toaster/tmpc/internal/util"
)

// Repository is a LyricsFetcher that delivers the lyrics of a song.
//
// @implements metadata.LyricsFetcher
type Repository struct{}

var _ metadata.LyricsFetcher = (*Repository)(nil)

// FetchLyrics returns the lyrics of a given song.
func (r *Repository) FetchLyrics(song *mpd.Song) ([]string, error) {
	if song == nil {
		return []string{}, nil
	}
	lyrics, err := r.findLyrics(song.Artist, song.Title)
	if err != nil && strings.Contains(song.Title, ":") {
		lyrics, err = r.findLyrics(song.Artist, strings.Split(song.Title, ":")[0])
	}
	if err != nil && strings.Contains(song.Title, "´") {
		lyrics, err = r.findLyrics(song.Artist, strings.ReplaceAll(song.Title, "´", "'"))
	}
	if err != nil && strings.Contains(song.Title, "") {
		lyrics, err = r.findLyrics(song.Artist, strings.ReplaceAll(song.Title, "’", "'"))
	}
	if err != nil {
		return nil, err
	}

	lines := []string{}
	var brDetected bool
	for c := lyrics.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			brDetected = false
			lines = append(lines, c.Data)
		case html.ElementNode:
			if c.Data == "br" {
				if brDetected {
					lines = append(lines, "")
				}
				brDetected = true
			}
		default:
			brDetected = false
		}
	}
	return lines, nil
}

func (r *Repository) findLyrics(artist, title string) (*html.Node, error) {
	url := fmt.Sprintf("https://lyrics.fandom.com/wiki/%s:%s", r.lyricsArg(artist), r.lyricsArg(title))
	doc, err := util.HTTPGetHTML(url)
	if err != nil {
		return nil, err
	}
	lyrics := r.findLyricsInHTML(doc)
	if lyrics == nil {
		altURL := r.findAltURLInHTML(doc)
		if altURL != "" {
			doc, err := util.HTTPGetHTML(fmt.Sprintf("https://lyrics.fandom.com%s", altURL))
			if err != nil {
				return nil, err
			}
			lyrics = r.findLyricsInHTML(doc)
		}
		if lyrics == nil {
			return nil, fmt.Errorf("could not find lyricbox in response from: %s", url)
		}
	}
	return lyrics, nil
}

func (r *Repository) findAltURLInHTML(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "span" {
		for _, a := range n.Attr {
			if a.Key == "class" && a.Val == "alternative-suggestion" {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode && c.Data == "a" {
						for _, aa := range c.Attr {
							if aa.Key == "href" {
								return aa.Val
							}
						}
					}
				}
				return ""
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		altURL := r.findAltURLInHTML(c)
		if altURL != "" {
			return altURL
		}
	}
	return ""
}

func (r *Repository) findLyricsInHTML(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "div" {
		for _, a := range n.Attr {
			if a.Key == "class" && a.Val == "lyricbox" {
				return n
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		lyrics := r.findLyricsInHTML(c)
		if lyrics != nil {
			return lyrics
		}
	}
	return nil
}

func (r *Repository) lyricsArg(s string) string {
	return strings.ReplaceAll(strings.Title(s), " ", "_")
}
