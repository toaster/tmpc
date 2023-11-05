package mpd

import "github.com/fhs/gompd/v2/mpd"

/*
Notes on MPD data:

Playlists: {
	"Last-Modified",
	"playlist",
}
*/

// Playlist represents an MPD playlist.
type Playlist struct {
	c     *Client
	name  string
	songs []*Song
}

// Name returns the name of the playlist.
func (p *Playlist) Name() string {
	return p.name
}

// Songs returns the songs contained by the playlist.
func (p *Playlist) Songs() ([]*Song, error) {
	if p.songs == nil {
		var rawSongs []mpd.Attrs
		err := p.c.retry(func() (err error) {
			rawSongs, err = p.c.c.PlaylistContents(p.name)
			return
		})
		if err != nil {
			return nil, err
		}
		p.songs = songsFromAttrs(rawSongs)
	}
	return p.songs, nil
}
