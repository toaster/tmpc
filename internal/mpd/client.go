package mpd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/fhs/gompd/mpd"
	"github.com/pkg/errors"
)

// TODO: AlbumYear im Song
// -> bedingt Abfrage des Albums? von MB?

// Client is a wrapper around a gompd client, providing high level functionality.
type Client struct {
	c            *mpd.Client
	invalidPerms bool
	onChange     func(string)
	onError      func(error)
	pass         string
	url          string
	w            *mpd.Watcher
}

// State is the current play state of the MPD server.
type State int

// State constants
const (
	StateStop State = iota
	StatePlay
	StatePause
)

// Status is detailed information on the currently played song.
type Status struct {
	Elapsed    int
	PlaylistID int
	State      State
	SongIdx    int
	SongID     int
}

// NewClient creates a new Client.
func NewClient(url, pass string, onChange func(string), onError func(error)) *Client {
	return &Client{url: url, pass: pass, onChange: onChange, onError: onError}
}

// AddToPlaylist adds the song to a playlist.
func (c *Client) AddToPlaylist(s *Song, name string) error {
	if c.c == nil {
		return nil
	}
	return c.retry(func() error { return c.c.PlaylistAdd(name, s.File) })
}

// AddToQueue adds the song to the queue and returns its ID.
func (c *Client) AddToQueue(s *Song, pos int) (id int, err error) {
	if c.c == nil {
		return 0, nil
	}
	err = c.retry(func() error {
		var e error
		id, e = c.c.AddID(s.File, pos)
		return e
	})
	return
}

// ClearQueue clears the queue.
func (c *Client) ClearQueue() error {
	if c.c == nil {
		return nil
	}
	return c.retry(func() error { return c.c.Clear() })
}

// Connect tries to connect to the MPD server.
func (c *Client) Connect() error {
	if c.invalidPerms {
		return errors.New("Insufficient MPD credentials: invalid permissions")
	}
	if c.c == nil {
		m, err := mpd.DialAuthenticated("tcp", c.url, c.pass)
		if err != nil {
			return err
		}
		c.c = m
	}
	if c.w == nil {
		w, err := mpd.NewWatcher("tcp", c.url, c.pass,
			"database", "options", "player", "playlist", "stored_playlist")
		if err != nil {
			return err
		}
		c.w = w
		go func() {
			for subsystem := range w.Event {
				c.onChange(subsystem)
			}
		}()
		go func() {
			for err2 := range w.Error {
				if e, ok := err2.(mpd.Error); ok && e.Code == mpd.ErrorPermission {
					if werr := c.w.Close(); werr != nil {
						c.onError(errors.Wrap(werr, "MPD watcher error"))
					}
					if cerr := c.c.Close(); cerr != nil {
						c.onError(errors.Wrap(cerr, "MPD client error"))
					}
					c.c = nil
					c.w = nil
					c.invalidPerms = true
				}
				c.onError(errors.Wrap(err2, "MPD watcher error"))
			}
		}()
	}
	return nil
}

// CurrentSongs delivers the songs of the queue.
func (c *Client) CurrentSongs() ([]*Song, error) {
	if c.c == nil {
		return nil, nil
	}
	var rawSongs []mpd.Attrs
	err := c.retry(func() (err error) {
		rawSongs, err = c.c.PlaylistInfo(-1, -1)
		return
	})
	if err != nil {
		return nil, err
	}
	return songsFromAttrs(rawSongs), nil
}

// Disconnect disconnects this client if possible.
func (c *Client) Disconnect() {
	if c.w != nil {
		c.w.Close()
		c.w = nil
	}
	if c.c != nil {
		c.c.Close()
		c.c = nil
	}
}

// IsConnected is true if the client is connected to the server.
func (c *Client) IsConnected() bool {
	return c.c != nil && c.w != nil
}

// Move moves a song inside the queue.
func (c *Client) Move(song *Song, index int) error {
	if c.c == nil {
		return nil
	}
	return c.retry(func() error { return c.c.MoveID(song.ID, index) })
}

// Next plays the next song of the queue.
func (c *Client) Next() error {
	if c.c == nil {
		return nil
	}
	return c.retry(func() error { return c.c.Next() })
}

// Pause pauses the playback.
func (c *Client) Pause(p bool) error {
	if c.c == nil {
		return nil
	}
	return c.retry(func() error { return c.c.Pause(p) })
}

func (c *Client) retry(f func() error) error {
	if err := f(); err != nil {
		c.c.Close()
		c.c, err = mpd.DialAuthenticated("tcp", c.url, c.pass)
		if err != nil {
			return err
		}
		return f()
	}
	return nil
}

// Play plays a specific song of the queue.
func (c *Client) Play(s *Song) error {
	if c.c == nil {
		return nil
	}
	return c.retry(func() error { return c.c.PlayID(s.ID) })
}

// PlayCurrent plays the current song of the queue.
func (c *Client) PlayCurrent() error {
	if c.c == nil {
		return nil
	}
	return c.retry(func() error { return c.c.Play(-1) })
}

// PlayID plays a specific song of the queue.
func (c *Client) PlayID(id int) error {
	if c.c == nil {
		return nil
	}
	return c.retry(func() error { return c.c.PlayID(id) })
}

// PlaylistLoad loads a playlist into the queue.
func (c *Client) PlaylistLoad(name string) error {
	if c.c == nil {
		return nil
	}
	return c.retry(func() error { return c.c.PlaylistLoad(name, 0, -1) })
}

// PlaylistRemove deletes the specified playlist.
func (c *Client) PlaylistRemove(name string) error {
	if c.c == nil {
		return nil
	}
	return c.retry(func() error { return c.c.PlaylistRemove(name) })
}

// Playlists delivers all playlists stored on the server.
func (c *Client) Playlists() ([]Playlist, error) {
	if c.c == nil {
		return nil, nil
	}
	var raw []mpd.Attrs
	err := c.retry(func() (err error) {
		raw, err = c.c.ListPlaylists()
		return
	})
	if err != nil {
		return nil, err
	}
	lists := make([]Playlist, len(raw))
	for i, attrs := range raw {
		name := attrs["playlist"]
		lists[i] = Playlist{c, name, nil}
	}
	return lists, nil
}

// Prev plays the previous song in the queue.
func (c *Client) Prev() error {
	if c.c == nil {
		return nil
	}
	return c.retry(func() error { return c.c.Previous() })
}

// RemoveFromQueue removes the given song from the queue.
func (c *Client) RemoveFromQueue(s *Song) error {
	if c.c == nil {
		return nil
	}
	return c.retry(func() error { return c.c.DeleteID(s.ID) })
}

// Search searches the MPD database.
func (c *Client) Search(key, value string, limit int) ([]*Song, error) {
	if c.c == nil {
		return nil, nil
	}
	var rawSongs []mpd.Attrs
	err := c.retry(func() (err error) {
		rawSongs, err = c.c.Search(key, value, "window", fmt.Sprintf("0:%d", limit))
		return
	})
	if err != nil {
		return nil, err
	}
	return songsFromAttrs(rawSongs), nil
}

// Seek seeks to a specific position in the currently played song.
func (c *Client) Seek(t int) error {
	if c.c == nil {
		return nil
	}
	return c.retry(func() error { return c.c.SeekCur(time.Duration(t)*time.Second, false) })
}

// Status delivers details on the currently played song.
func (c *Client) Status() (*Status, error) {
	if c.c == nil {
		return &Status{}, nil
	}
	var attrs mpd.Attrs
	err := c.retry(func() (err error) {
		attrs, err = c.c.Status()
		return
	})
	if err != nil {
		return nil, err
	}
	var elapsed int
	fmt.Sscanf(attrs["elapsed"], "%d", &elapsed)
	plID, _ := strconv.Atoi(attrs["playlist"])
	sID, _ := strconv.Atoi(attrs["songid"])
	sIdx, _ := strconv.Atoi(attrs["song"])
	var state State
	switch attrs["state"] {
	case "play":
		state = StatePlay
	case "pause":
		state = StatePause
	default:
		state = StateStop
	}
	return &Status{
		Elapsed:    elapsed,
		PlaylistID: plID,
		State:      state,
		SongID:     sID,
		SongIdx:    sIdx,
	}, nil
}

// Stop stops the playback.
func (c *Client) Stop() error {
	if c.c == nil {
		return nil
	}
	return c.retry(func() error { return c.c.Stop() })
}
