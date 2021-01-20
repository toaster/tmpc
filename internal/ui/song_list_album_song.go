package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/toaster/tmpc/internal/mpd"
)

var _ desktop.Hoverable = (*songListAlbumSong)(nil)
var _ desktop.Mouseable = (*songListAlbumSong)(nil)
var _ fyne.DoubleTappable = (*songListAlbumSong)(nil)
var _ fyne.Draggable = (*songListAlbumSong)(nil)
var _ fyne.Tappable = (*songListAlbumSong)(nil)

type songListAlbumSong struct {
	*songListAlbumLine
	contextMenu       *fyne.Menu
	endDrag           func()
	isDragging        func() bool
	onMark            func(desktop.Modifier, *songListAlbumSong)
	onClick           func(*mpd.Song)
	setDragMarkAfter  func(*mpd.Song, bool)
	setDragMarkBefore func(*mpd.Song, bool)
	startDrag         func()
	song              *mpd.Song
}

func newSongListAlbumSong(contextMenu *fyne.Menu, song *mpd.Song, insertSelection func(), selectionIsDragged func() bool, markSong func(desktop.Modifier, *songListAlbumSong), onSongClick func(*mpd.Song), coverSize float32, setDragMarkAfter func(*mpd.Song, bool), setDragMarkBefore func(*mpd.Song, bool), dragSelection func()) *songListAlbumSong {
	s := &songListAlbumSong{
		contextMenu: contextMenu,
		endDrag:     insertSelection,
		isDragging:  selectionIsDragged,
		onMark:      markSong,
		onClick:     onSongClick,
		songListAlbumLine: &songListAlbumLine{
			lastTextRight: true,
			pad:           coverSize,
			texts: []string{
				fmt.Sprintf("%02d", song.Track),
				song.DisplayTitle(),
				timeString(song.Time),
			},
		},
		setDragMarkAfter:  setDragMarkAfter,
		setDragMarkBefore: setDragMarkBefore,
		song:              song,
		startDrag:         dragSelection,
	}
	s.ExtendBaseWidget(s)
	return s
}

func (s *songListAlbumSong) CreateRenderer() fyne.WidgetRenderer {
	return s.songListAlbumLine.CreateRenderer()
}

func (s *songListAlbumSong) DoubleTapped(_ *fyne.PointEvent) {
	s.onClick(s.song)
}

func (s *songListAlbumSong) DragEnd() {
	s.endDrag()
}

func (s *songListAlbumSong) Dragged(*fyne.DragEvent) {
	if !s.isDragging() {
		s.startDrag()
	}
}

func (s *songListAlbumSong) MouseDown(e *desktop.MouseEvent) {
	s.onMark(e.Modifier, s)
}

func (s *songListAlbumSong) MouseIn(e *desktop.MouseEvent) {
	s.hovered = true
	if s.isDragging() {
		s.showInsertMarker = true
		s.setDragMark(e.Position, true)
	}
	// TODO: dieses ganze refresh-Geraffel ist hochgradig undurchsichtig
	// - normalerweise reicht canvas.Refresh(w) aber das triggert keinen <widget>.Refresh
	// - <widget>.Refresh soll wahrscheinlich aktualisieren (show/hide o.ä.) ohne layout/minsize-Änderung.
	// - <widget>.Refresh muss aber explizit gerufen werden und hängt damit von widget.Renderer() ab,
	//   was ja nun ganz und gar nicht geht.
	// - ausserdem is offenbar ein canvas.Refresh(renderer.widget) im Renderer gefährlich, wenn dessen widget embedded ist
	s.Refresh()
	canvas.Refresh(s)
}

func (s *songListAlbumSong) MouseMoved(e *desktop.MouseEvent) {
	if s.isDragging() {
		if s.setDragMark(e.Position, false) {
			// TODO: s.o.
			s.Refresh()
			canvas.Refresh(s)
		}
	}
}

func (s *songListAlbumSong) MouseOut() {
	s.showInsertMarker = false
	s.hovered = false
	// TODO s.o.
	s.Refresh()
	canvas.Refresh(s)
}

func (s *songListAlbumSong) MouseUp(*desktop.MouseEvent) {
}

func (s *songListAlbumSong) Tapped(_ *fyne.PointEvent) {
}

func (s *songListAlbumSong) TappedSecondary(e *fyne.PointEvent) {
	c := fyne.CurrentApp().Driver().CanvasForObject(s)
	widget.ShowPopUpMenuAtPosition(s.contextMenu, c, e.AbsolutePosition)
}

func (s *songListAlbumSong) setDragMark(p fyne.Position, force bool) bool {
	bottom := p.Y > 10
	if (bottom != s.insertMarkerBottom) || force {
		s.insertMarkerBottom = bottom
		if bottom {
			s.setDragMarkAfter(s.song, s.selected)
		} else {
			s.setDragMarkBefore(s.song, s.selected)
		}
		return true
	}
	return false
}
