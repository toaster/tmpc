package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/toaster/tmpc/internal/mpd"
)

var _ desktop.Hoverable = (*songListAlbum)(nil)
var _ desktop.Mouseable = (*songListAlbum)(nil)
var _ fyne.Draggable = (*songListAlbum)(nil)
var _ fyne.Tappable = (*songListAlbum)(nil)

type songListAlbum struct {
	widget.BaseWidget
	contextMenu       *fyne.Menu
	cover             *songListAlbumCover
	endDrag           func()
	indicatedSong     int
	indicator         fyne.CanvasObject
	isDragging        func() bool
	header            *songListAlbumLine
	markAlbum         func(desktop.Modifier, *songListAlbum)
	setDragMarkAfter  func(*mpd.Song, bool)
	setDragMarkBefore func(*mpd.Song, bool)
	showSongs         bool
	songs             []*songListAlbumSong
	startDrag         func()
	summary           *songListAlbumLine
}

func newSongListAlbum(
	songs []*mpd.Song,
	markSong func(desktop.Modifier, *songListAlbumSong),
	markAlbum func(desktop.Modifier, *songListAlbum),
	dragSelection func(),
	selectionIsDragged func() bool,
	setDragMarkBefore, setDragMarkAfter func(*mpd.Song, bool),
	insertSelection func(),
	onSongClick func(*mpd.Song),
	contextMenu *fyne.Menu,
) *songListAlbum {
	album := &songListAlbum{
		contextMenu:       contextMenu,
		indicatedSong:     -1,
		endDrag:           insertSelection,
		isDragging:        selectionIsDragged,
		markAlbum:         markAlbum,
		setDragMarkAfter:  setDragMarkAfter,
		setDragMarkBefore: setDragMarkBefore,
		showSongs:         true,
		startDrag:         dragSelection,
	}
	album.ExtendBaseWidget(album)
	coverSize :=
		// +1 is to ajust the textsize/dip difference
		// would be nice to have a text size to dip coverter
		(theme.TextSize()+1)*2 +
			// 3 times padding is the bottom padding of the head line and the whole padding of the first line
			// +1 is the separator between header and first line
			3*theme.Padding() + 1
	time := 0
	album.songs = make([]*songListAlbumSong, 0, len(songs))
	for _, song := range songs {
		entry := newSongListAlbumSong(contextMenu, song, insertSelection, selectionIsDragged, markSong, onSongClick, coverSize, setDragMarkAfter, setDragMarkBefore, dragSelection)
		album.songs = append(album.songs, entry)
		time += song.Time
	}
	album.header = newAlbumHeadLine(coverSize, []string{
		fmt.Sprintf("%s - %s (%d)", songs[0].AlbumArtist, songs[0].Album, songs[0].Year),
		timeString(time),
	})
	album.summary = newAlbumEntryLine(coverSize, []string{fmt.Sprintf("%d", len(songs)), "Tracks"})
	img := canvas.NewImageFromResource(nil)
	img.SetMinSize(fyne.NewSize(coverSize, coverSize))
	album.cover = newSongListAlbumCover(img, func() {
		album.showSongs = !album.showSongs
		album.Refresh()
	})

	return album
}

func (a *songListAlbum) CreateRenderer() fyne.WidgetRenderer {
	gradient := canvas.NewHorizontalGradient(theme.BackgroundColor(), nil)
	objects := make([]fyne.CanvasObject, 0, len(a.songs)+5)
	for _, s := range a.songs {
		objects = append(objects, s)
	}
	objects = append(objects, a.header, a.summary, a.cover, gradient)
	return &songListAlbumRenderer{
		a: a,
		baseRenderer: baseRenderer{
			objects: objects,
		},
		gradient:            gradient,
		objectsButIndicator: objects,
	}
}

func (a *songListAlbum) DragEnd() {
	a.endDrag()
}

func (a *songListAlbum) Dragged(*fyne.DragEvent) {
	if !a.isDragging() {
		a.startDrag()
	}
}

func (a *songListAlbum) MouseDown(e *desktop.MouseEvent) {
	a.markAlbum(e.Modifier, a)
}

func (a *songListAlbum) MouseIn(e *desktop.MouseEvent) {
	a.header.hovered = true
	a.summary.hovered = true
	if a.isDragging() {
		a.setDragMark(&e.PointEvent, true)
	}

	a.header.Refresh()
	a.summary.Refresh()
}

func (a *songListAlbum) MouseMoved(e *desktop.MouseEvent) {
	if a.isDragging() {
		a.setDragMark(&e.PointEvent, false)
	}
	a.header.Refresh()
	a.summary.Refresh()
}

func (a *songListAlbum) MouseOut() {
	a.header.hovered = false
	a.header.showInsertMarker = false
	a.summary.hovered = false
	a.summary.showInsertMarker = false
	a.header.Refresh()
	a.summary.Refresh()
}

func (a *songListAlbum) MouseUp(*desktop.MouseEvent) {
}

func (a *songListAlbum) Tapped(_ *fyne.PointEvent) {
}

func (a *songListAlbum) TappedSecondary(e *fyne.PointEvent) {
	c := fyne.CurrentApp().Driver().CanvasForObject(a)
	widget.ShowPopUpMenuAtPosition(a.contextMenu, c, e.AbsolutePosition)
}

// UpdateCover changes the cover.
func (a *songListAlbum) UpdateCover(cover fyne.Resource) {
	a.cover.Update(cover)
}

func (a *songListAlbum) setSongIndicator(idx int, indicator fyne.CanvasObject) {
	a.indicator = indicator
	a.indicatedSong = idx
	a.Refresh()
}

func (a *songListAlbum) setDragMark(e *fyne.PointEvent, force bool) {
	if eventIsOn(e, a.header) {
		if !a.header.showInsertMarker || force {
			a.header.showInsertMarker = true
			a.summary.showInsertMarker = false
			s := a.songs[0]
			a.setDragMarkBefore(s.song, s.selected)
			a.header.Refresh()
			a.summary.Refresh()
		}
	} else {
		if !a.summary.showInsertMarker || force {
			a.header.showInsertMarker = false
			a.summary.showInsertMarker = true
			s := a.songs[len(a.songs)-1]
			a.setDragMarkAfter(s.song, s.selected)
			a.header.Refresh()
			a.summary.Refresh()
		}
	}
}

func (a *songListAlbum) removeSongIndicator() {
	a.indicatedSong = -1
	a.indicator = nil
	a.Refresh()
}

type songListAlbumRenderer struct {
	baseRenderer
	a                   *songListAlbum
	gradient            *canvas.LinearGradient
	minSize             fyne.Size
	objectsButIndicator []fyne.CanvasObject
}

func (r *songListAlbumRenderer) Layout(size fyne.Size) {
	if r.a.showSongs {
		r.a.summary.Hide()
		for _, e := range r.a.songs {
			e.Show()
		}
	} else {
		r.a.summary.Show()
		for _, e := range r.a.songs {
			e.Hide()
		}
	}

	if (r.a.header.Position() == fyne.Position{}) {
		r.a.cover.Resize(r.a.cover.MinSize())
		r.a.cover.Move(fyne.NewPos(theme.Padding(), 0))
		r.gradient.Move(fyne.NewPos(theme.Padding(), r.a.cover.MinSize().Height))
		r.a.header.Move(fyne.NewPos(theme.Padding(), 0))
		yOffset := r.a.header.MinSize().Height
		r.a.summary.Move(fyne.NewPos(theme.Padding(), yOffset))
		for _, e := range r.a.songs {
			e.Move(fyne.NewPos(theme.Padding(), yOffset))
			yOffset += e.MinSize().Height
		}
	}
	r.gradient.Resize(fyne.NewSize(r.a.cover.MinSize().Width, size.Height-r.a.cover.MinSize().Height))
	r.a.header.Resize(fyne.NewSize(size.Width-2*theme.Padding(), r.a.header.MinSize().Height))
	r.a.summary.Resize(fyne.NewSize(size.Width-2*theme.Padding(), r.a.summary.MinSize().Height))
	for _, e := range r.a.songs {
		e.Resize(fyne.NewSize(size.Width-2*theme.Padding(), e.MinSize().Height))
	}
}

func (r *songListAlbumRenderer) MinSize() fyne.Size {
	if (r.minSize == fyne.Size{}) {
		r.computeMinSize()
	}
	return r.minSize
}

func (r *songListAlbumRenderer) Refresh() {
	sc := theme.BackgroundColor()
	if r.gradient.StartColor != sc {
		r.gradient.StartColor = sc
		r.Refresh()
	}
	if r.a.indicatedSong > -1 {
		var linePos fyne.Position
		if r.a.showSongs {
			linePos = r.a.songs[r.a.indicatedSong].Position()
		} else {
			linePos = r.a.summary.Position()
		}
		r.objects = append(r.objectsButIndicator, r.a.indicator)
		pos := fyne.NewPos(r.a.cover.MinSize().Width-18, linePos.Y+theme.Padding()/2)
		r.a.indicator.Move(pos)
	} else {
		r.objects = r.objectsButIndicator
	}
	r.computeMinSize()
	canvas.Refresh(r.a)
}

func (r *songListAlbumRenderer) computeMinSize() {
	minWidth := r.a.header.MinSize().Width
	minHeight := r.a.header.MinSize().Height
	if r.a.showSongs {
		for _, s := range r.a.songs {
			minWidth = fyne.Max(minWidth, s.MinSize().Width)
			minHeight += s.MinSize().Height
		}
	} else {
		minWidth = fyne.Max(minWidth, r.a.summary.MinSize().Width)
		minHeight += r.a.summary.MinSize().Height
	}
	r.minSize = fyne.NewSize(minWidth+r.a.cover.MinSize().Width+2*theme.Padding(), minHeight)
}

func timeString(seconds int) string {
	h := seconds / 3600
	m := seconds % 3600 / 60
	s := seconds % 60
	switch {
	case h > 0:
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	default:
		return fmt.Sprintf("%d:%02d", m, s)
	}
}
