package ui

import (
	"fyne.io/fyne"
	"fyne.io/fyne/widget"

	"github.com/toaster/tmpc/internal/mpd"
)

var _ fyne.Widget = (*PlaylistList)(nil)

// PlaylistList is a container showing MPD playlists.
type PlaylistList struct {
	widget.BaseWidget
	box     *widget.Box
	delete  func(string)
	playNow func(string)
}

// NewPlaylistList returns a new playlist container.
func NewPlaylistList(playNow, delete func(string)) *PlaylistList {
	l := &PlaylistList{
		box:     widget.NewVBox(),
		delete:  delete,
		playNow: playNow,
	}
	l.ExtendBaseWidget(l)
	return l
}

// CreateRenderer is an internal method
func (p *PlaylistList) CreateRenderer() fyne.WidgetRenderer {
	return p.box.CreateRenderer()
}

// Update replaces the lists contents with new ones.
func (p *PlaylistList) Update(pls []mpd.Playlist) {
	p.box.Children = make([]fyne.CanvasObject, 0, len(pls))
	for _, pl := range pls {
		p.box.Children = append(p.box.Children, newPlaylistListEntry(pl.Name(), p.playNow, p.delete))
	}
	p.Refresh()
}
