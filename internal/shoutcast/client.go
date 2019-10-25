package shoutcast

import (
	"io"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
	"github.com/romantomjak/shoutcast"
)

type Client struct {
	context *oto.Context
	decoder *mp3.Decoder
	onError func(error)
	player  *oto.Player
	playing bool
	stream  *shoutcast.Stream
	url     string
}

func NewClient(url string, onError func(error)) *Client {
	return &Client{onError: onError, url: url}
}

func (c *Client) IsPlaying() bool {
	return c.playing
}

func (c *Client) Play() error {
	if c.playing {
		return nil
	}

	var err error
	c.stream, err = shoutcast.Open(c.url)
	if err != nil {
		return err
	}

	c.decoder, err = mp3.NewDecoder(c.stream)
	if err != nil {
		c.stream.Close()
		return err
	}

	c.context, err = oto.NewContext(c.decoder.SampleRate(), 2, 2, 8192)
	if err != nil {
		c.stream.Close()
		return err
	}
	c.player = c.context.NewPlayer()

	c.playing = true
	go func() {
		defer c.context.Close()
		defer c.stream.Close()
		defer func() { c.playing = false }()

		if _, err := io.Copy(c.player, c.decoder); err != nil {
			// ignore errors that occur because Stop() has been called
			if c.playing {
				c.onError(err)
			}
		}
	}()

	return nil
}

func (c *Client) Stop() {
	if !c.playing {
		return
	}
	c.playing = false
	// Don't close the context here, or the copying go routine won't stop.
	c.stream.Close()
}
