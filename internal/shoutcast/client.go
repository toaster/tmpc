package shoutcast

import (
	"time"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v2"
	"github.com/romantomjak/shoutcast"
)

// Client is a ShoutCast client.
type Client struct {
	context *oto.Context
	decoder *mp3.Decoder
	onError func(error)
	player  oto.Player
	stop    chan bool
	stream  *shoutcast.Stream
	url     string
}

// NewClient creates a new Client for a URL.
// `onError` is a function which receives all occurring errors.
func NewClient(url string, onError func(error)) *Client {
	return &Client{
		onError: onError,
		stop:    make(chan bool),
		url:     url,
	}
}

// IsPlaying returns `true` iff the client is currently playing.
func (c *Client) IsPlaying() bool {
	return c.player != nil && c.player.IsPlaying()
}

// Play starts to play back the ShoutCast stream at the client’s URL.
func (c *Client) Play() error {
	if c.IsPlaying() {
		return nil
	}

	var err error
	c.stream, err = shoutcast.Open(c.url)
	if err != nil {
		return err
	}

	c.decoder, err = mp3.NewDecoder(c.stream)
	if err != nil {
		_ = c.stream.Close()
		return err
	}

	var ready chan struct{}
	c.context, ready, err = oto.NewContext(c.decoder.SampleRate(), 2, 2)
	if err != nil {
		_ = c.stream.Close()
		return err
	}
	go func() {
		<-ready

		c.player = c.context.NewPlayer(c.decoder)
		c.player.Play()

		defer func() { _ = c.player.Close() }()
		defer func() { _ = c.stream.Close() }()

		var stop bool
		for c.player.IsPlaying() && !stop {
			select {
			case <-c.stop:
				stop = true
			case <-time.After(1 * time.Second):
			}
		}
	}()

	return nil
}

// Stop stops the playback.
func (c *Client) Stop() {
	if !c.IsPlaying() {
		return
	}

	go func() {
		c.stop <- true
	}()
}
