package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

var _ desktop.Hoverable = (*playlistListEntry)(nil)
var _ fyne.Tappable = (*playlistListEntry)(nil)

type playlistListEntry struct {
	listEntry
	contextMenu *fyne.Menu
	name        string
}

func newPlaylistListEntry(name string, playNow, deletePL func(string), w fyne.Window) *playlistListEntry {
	remove := func() {
		callback := func(confirmed bool) {
			if confirmed {
				deletePL(name)
			}
		}
		dialog.ShowConfirm(
			"Delete Playlist",
			fmt.Sprintf("Do you really want to remove the playlist “%s”?\nThis cannot be undone.", name),
			callback,
			w,
		)
	}
	items := []*fyne.MenuItem{
		fyne.NewMenuItem("Delete", remove),
		fyne.NewMenuItem("Play Now And Replace Queue", func() { playNow(name) }),
	}
	e := &playlistListEntry{contextMenu: fyne.NewMenu("", items...), name: name}
	e.ExtendBaseWidget(e)
	return e
}

func (e *playlistListEntry) CreateRenderer() fyne.WidgetRenderer {
	text := widget.NewLabel(e.name)
	ler := e.listEntry.createRenderer()
	ler.objects = append(ler.objects, text)
	return &playlistListEntryRenderer{
		listEntryRenderer: ler,
		e:                 e,
		text:              text,
	}
}

func (e *playlistListEntry) MouseIn(_ *desktop.MouseEvent) {
	e.hovered = true
	e.Refresh()
}

func (*playlistListEntry) MouseMoved(_ *desktop.MouseEvent) {
}

func (e *playlistListEntry) MouseOut() {
	e.hovered = false
	e.Refresh()
}

func (e *playlistListEntry) Refresh() {
	// TODO: widget extension + WidgetRenderer + refreshing is still error-prone
	e.listEntry.Refresh()
	canvas.Refresh(e)
}

func (*playlistListEntry) Tapped(_ *fyne.PointEvent) {
}

func (e *playlistListEntry) TappedSecondary(pe *fyne.PointEvent) {
	c := fyne.CurrentApp().Driver().CanvasForObject(e)
	widget.ShowPopUpMenuAtPosition(e.contextMenu, c, pe.AbsolutePosition)
}

type playlistListEntryRenderer struct {
	listEntryRenderer
	e    *playlistListEntry
	text fyne.CanvasObject
}

func (r *playlistListEntryRenderer) Layout(size fyne.Size) {
	r.text.Resize(r.text.MinSize())
	r.listEntryRenderer.Layout(size)
}

func (r *playlistListEntryRenderer) MinSize() fyne.Size {
	return r.text.MinSize().Add(r.listEntryRenderer.MinSize())
}
