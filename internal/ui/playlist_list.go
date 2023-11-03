package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/toaster/tmpc/internal/mpd"
)

var _ fyne.Widget = (*PlaylistList)(nil)

// PlaylistList is a container showing MPD playlists.
type PlaylistList struct {
	widget.BaseWidget
	box     *fyne.Container
	delete  func(string)
	playNow func(string)
	window  fyne.Window
}

// NewPlaylistList returns a new playlist container.
func NewPlaylistList(playNow, delete func(string), w fyne.Window) *PlaylistList {
	l := &PlaylistList{
		box:     container.NewVBox(),
		delete:  delete,
		playNow: playNow,
		window:  w,
	}
	l.ExtendBaseWidget(l)
	return l
}

// CreateRenderer is an internal method
func (p *PlaylistList) CreateRenderer() fyne.WidgetRenderer {
	return &containerRenderer{p.box}
}

// Update replaces the lists contents with new ones.
func (p *PlaylistList) Update(pls []*mpd.Playlist) {
	p.box.Objects = make([]fyne.CanvasObject, 0, len(pls))
	for _, pl := range pls {
		p.box.Objects = append(p.box.Objects, newPlaylistListEntry(pl.Name(), p.playNow, p.delete, p.window))
	}
	p.Refresh()
}
