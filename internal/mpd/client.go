package mpd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/fhs/gompd/mpd"
	"github.com/pkg/errors"
)

// TODO: Errorhandling für alle MPD-Calls via registriertem Error-Handler,
// welcher dann in der App einen Dialog anzeigt???
// Die Methoden müssen dann einen geeigneten Leerwert (wahrscheinlich reicht der Nullwert)
// zurückliefern.

// TODO: AlbumYear im Song
// -> bedingt Abfrage des Albums? von MB?

// Client is a wrapper around a gompd client, providing high level functionality.
type Client struct {
	c        *mpd.Client
	onChange func(string)
	onError  func(error)
	pass     string
	url      string
	w        *mpd.Watcher
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

// ClearQueue clears the queue.
func (c *Client) ClearQueue() error {
	if c.c == nil {
		return nil
	}
	return c.retry(func() error { return c.c.Clear() })
}

// Connect tries to connect to the MPD server.
func (c *Client) Connect() error {
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
			for err := range w.Error {
				c.onError(errors.Wrap(err, "MPD Watcher error"))
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
