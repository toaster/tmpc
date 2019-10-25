package ui

import (
	"fmt"
	"image"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/toaster/tmpc/internal/mpd"
)

var _ desktop.Hoverable = (*queueAlbum)(nil)
var _ desktop.Mouseable = (*queueAlbum)(nil)
var _ fyne.Draggable = (*queueAlbum)(nil)

type queueAlbum struct {
	baseWidget
	cover             fyne.CanvasObject
	currentSong       int
	endDrag           func()
	isDragging        func() bool
	header            *queueAlbumLine
	markAlbum         func(desktop.Modifier, *queueAlbum)
	playState         PlayerState
	setDragMarkAfter  func(*mpd.Song, bool)
	setDragMarkBefore func(*mpd.Song, bool)
	showSongs         bool
	songs             []*queueAlbumSong
	startDrag         func()
	summary           *queueAlbumLine
}

func newQueueAlbum(
	coverImage image.Image,
	songs []*mpd.Song,
	markSong func(desktop.Modifier, *queueAlbumSong),
	markAlbum func(desktop.Modifier, *queueAlbum),
	dragSelection func(),
	selectionIsDragged func() bool,
	setDragMarkBefore, setDragMarkAfter func(*mpd.Song, bool),
	insertSelection func(),
	playNow, removeFromQueue, removeOthersFromQueue, removeBeforeFromQueue, removeAfterFromQueue func(*mpd.Song),
	clearQueue func(),
) *queueAlbum {
	album := &queueAlbum{
		currentSong:       -1,
		endDrag:           insertSelection,
		isDragging:        selectionIsDragged,
		markAlbum:         markAlbum,
		playState:         PlayerStateStop,
		setDragMarkAfter:  setDragMarkAfter,
		setDragMarkBefore: setDragMarkBefore,
		showSongs:         true,
		startDrag:         dragSelection,
	}
	coverSize :=
		// +1 is to ajust the textsize/dip difference
		// would be nice to have a text size to dip coverter
		(theme.TextSize()+1)*2 +
			// 3 times padding is the bottom padding of the head line and the whole padding of the first line
			// +1 is the separator between header and first line
			3*theme.Padding() + 1
	time := 0
	album.songs = make([]*queueAlbumSong, 0, len(songs))
	for _, song := range songs {
		entry := &queueAlbumSong{
			clearQueue: clearQueue,
			endDrag:    insertSelection,
			isDragging: selectionIsDragged,
			onClick:    markSong,
			playNow:    playNow,
			queueAlbumLine: &queueAlbumLine{
				lastTextRight: true,
				pad:           coverSize,
				texts: []string{
					fmt.Sprintf("%02d", song.Track),
					song.Title,
					timeString(song.Time),
				},
			},
			removeAfterFromQueue:  removeAfterFromQueue,
			removeBeforeFromQueue: removeBeforeFromQueue,
			removeFromQueue:       removeFromQueue,
			removeOthersFromQueue: removeOthersFromQueue,
			setDragMarkAfter:      setDragMarkAfter,
			setDragMarkBefore:     setDragMarkBefore,
			song:                  song,
			startDrag:             dragSelection,
		}
		album.songs = append(album.songs, entry)
		time += song.Time
	}
	album.header = &queueAlbumLine{
		bold:          true,
		lastTextRight: true,
		pad:           coverSize,
		texts: []string{
			fmt.Sprintf("%s - %s (%d)", songs[0].AlbumArtist, songs[0].Album, songs[0].Year),
			timeString(time),
		},
	}
	album.summary = &queueAlbumLine{
		insertMarkerBottom: true,
		pad:                coverSize,
		texts:              []string{fmt.Sprintf("%d", len(songs)), "Tracks"},
	}
	img := canvas.NewImageFromImage(coverImage)
	img.SetMinSize(fyne.NewSize(coverSize, coverSize))
	album.cover = &queueAlbumCover{
		image: img,
		onClick: func() {
			album.showSongs = !album.showSongs
			widget.Refresh(album)
		},
	}

	return album
}

func (a *queueAlbum) CreateRenderer() fyne.WidgetRenderer {
	gradient := canvas.NewHorizontalGradient(theme.BackgroundColor(), nil)
	objects := make([]fyne.CanvasObject, 0, len(a.songs)+4)
	for _, s := range a.songs {
		objects = append(objects, s)
	}
	playIndicator := canvas.NewImageFromResource(rscPlayIndicator)
	playIndicator.Hide()
	playIndicator.Resize(fyne.NewSize(18, 18))
	pauseIndicator := canvas.NewImageFromResource(rscPauseIndicator)
	pauseIndicator.Hide()
	pauseIndicator.Resize(fyne.NewSize(18, 18))
	stopIndicator := canvas.NewImageFromResource(rscStopIndicator)
	stopIndicator.Hide()
	stopIndicator.Resize(fyne.NewSize(18, 18))
	objects = append(objects, a.header, a.summary, a.cover, gradient, playIndicator, pauseIndicator, stopIndicator)
	return &queueAlbumRenderer{
		a: a,
		baseRenderer: baseRenderer{
			objects: objects,
		},
		gradient:       gradient,
		playIndicator:  playIndicator,
		pauseIndicator: pauseIndicator,
		stopIndicator:  stopIndicator,
	}
}

func (a *queueAlbum) DragEnd() {
	a.endDrag()
}

func (a *queueAlbum) Dragged(e *fyne.DragEvent) {
	if !a.isDragging() {
		a.startDrag()
	}
}

func (a *queueAlbum) Hide() {
	a.hide(a)
}

func (a *queueAlbum) MinSize() fyne.Size {
	return a.minSize(a)
}

func (a *queueAlbum) MouseDown(e *desktop.MouseEvent) {
	a.markAlbum(e.Modifier, a)
}

func (a *queueAlbum) MouseIn(e *desktop.MouseEvent) {
	a.header.hovered = true
	a.summary.hovered = true
	if a.isDragging() {
		a.setDragMark(&e.PointEvent, true)
	}

	canvas.Refresh(a.header)
	canvas.Refresh(a.summary)
}

func (a *queueAlbum) MouseMoved(e *desktop.MouseEvent) {
	if a.isDragging() {
		a.setDragMark(&e.PointEvent, false)
	}
	canvas.Refresh(a.header)
	canvas.Refresh(a.summary)
}

func (a *queueAlbum) MouseOut() {
	a.header.hovered = false
	a.header.showInsertMarker = false
	a.summary.hovered = false
	a.summary.showInsertMarker = false
	// TODO s. queue_album_song
	widget.Renderer(a.header).Refresh()
	widget.Renderer(a.summary).Refresh()
	canvas.Refresh(a.header)
	canvas.Refresh(a.summary)
}

func (a *queueAlbum) MouseUp(e *desktop.MouseEvent) {
}

func (a *queueAlbum) Resize(size fyne.Size) {
	a.resize(a, size)
}

func (a *queueAlbum) setCurrentSong(idx int, ps PlayerState) {
	a.playState = ps
	a.currentSong = idx
	widget.Refresh(a)
}

func (a *queueAlbum) Show() {
	a.show(a)
}

func (a *queueAlbum) setDragMark(e *fyne.PointEvent, force bool) {
	if eventIsOn(e, a.header) {
		if a.header.showInsertMarker == false || force {
			a.header.showInsertMarker = true
			a.summary.showInsertMarker = false
			s := a.songs[0]
			a.setDragMarkBefore(s.song, s.selected)
			// TODO s. queue_album_song
			widget.Renderer(a.header).Refresh()
			widget.Renderer(a.summary).Refresh()
		}
	} else {
		if a.summary.showInsertMarker == false || force {
			a.header.showInsertMarker = false
			a.summary.showInsertMarker = true
			s := a.songs[len(a.songs)-1]
			a.setDragMarkAfter(s.song, s.selected)
			// TODO s. queue_album_song
			widget.Renderer(a.header).Refresh()
			widget.Renderer(a.summary).Refresh()
		}
	}
}

func (a *queueAlbum) unsetCurrentSong() {
	a.currentSong = -1
	widget.Refresh(a)
}

type queueAlbumRenderer struct {
	baseRenderer
	a              *queueAlbum
	gradient       fyne.CanvasObject
	minSize        fyne.Size
	playIndicator  fyne.CanvasObject
	pauseIndicator fyne.CanvasObject
	stopIndicator  fyne.CanvasObject
}

func (r *queueAlbumRenderer) Layout(size fyne.Size) {
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

func (r *queueAlbumRenderer) MinSize() fyne.Size {
	if (r.minSize == fyne.Size{}) {
		r.computeMinSize()
	}
	return r.minSize
}

func (r *queueAlbumRenderer) Refresh() {
	if r.a.currentSong > -1 {
		r.playIndicator.Hide()
		r.pauseIndicator.Hide()
		r.stopIndicator.Hide()
		switch r.a.playState {
		case PlayerStatePlay:
			r.playIndicator.Show()
		case PlayerStatePause:
			r.pauseIndicator.Show()
		case PlayerStateStop:
			r.stopIndicator.Show()
		}
		var linePos fyne.Position
		if r.a.showSongs {
			linePos = r.a.songs[r.a.currentSong].Position()
		} else {
			linePos = r.a.summary.Position()
		}
		pos := fyne.NewPos(r.a.cover.MinSize().Width-18, linePos.Y+theme.Padding()/2)
		if pos != r.playIndicator.Position() {
			r.playIndicator.Move(pos)
			r.pauseIndicator.Move(pos)
			r.stopIndicator.Move(pos)
		}
		canvas.Refresh(r.playIndicator)
		canvas.Refresh(r.pauseIndicator)
		canvas.Refresh(r.stopIndicator)
	} else {
		if r.playIndicator.Visible() {
			r.playIndicator.Hide()
			canvas.Refresh(r.playIndicator)
		}
		if r.pauseIndicator.Visible() {
			r.pauseIndicator.Hide()
			canvas.Refresh(r.pauseIndicator)
		}
		if r.stopIndicator.Visible() {
			r.stopIndicator.Hide()
			canvas.Refresh(r.stopIndicator)
		}
	}
	r.computeMinSize()
	canvas.Refresh(r.a)
}

func (r *queueAlbumRenderer) computeMinSize() {
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
