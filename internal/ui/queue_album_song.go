package ui

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/widget"
	"github.com/toaster/tmpc/internal/mpd"
)

var _ desktop.Hoverable = (*queueAlbumSong)(nil)
var _ desktop.Mouseable = (*queueAlbumSong)(nil)
var _ fyne.DoubleTappable = (*queueAlbumSong)(nil)
var _ fyne.Draggable = (*queueAlbumSong)(nil)
var _ fyne.Tappable = (*queueAlbumSong)(nil)

type queueAlbumSong struct {
	*queueAlbumLine
	clearQueue            func()
	endDrag               func()
	isDragging            func() bool
	onClick               func(desktop.Modifier, *queueAlbumSong)
	playNow               func(*mpd.Song)
	removeFromQueue       func(*mpd.Song)
	removeOthersFromQueue func(*mpd.Song)
	removeBeforeFromQueue func(*mpd.Song)
	removeAfterFromQueue  func(*mpd.Song)
	setDragMarkAfter      func(*mpd.Song, bool)
	setDragMarkBefore     func(*mpd.Song, bool)
	startDrag             func()
	song                  *mpd.Song
}

func (s *queueAlbumSong) CreateRenderer() fyne.WidgetRenderer {
	return widget.Renderer(s.queueAlbumLine)
}

func (s *queueAlbumSong) DoubleTapped(_ *fyne.PointEvent) {
	s.playNow(s.song)
}

func (s *queueAlbumSong) DragEnd() {
	s.endDrag()
}

func (s *queueAlbumSong) Dragged(e *fyne.DragEvent) {
	if !s.isDragging() {
		s.startDrag()
	}
}

func (s *queueAlbumSong) Hide() {
	s.hide(s)
}

func (s *queueAlbumSong) MinSize() fyne.Size {
	return s.minSize(s)
}

func (s *queueAlbumSong) MouseDown(e *desktop.MouseEvent) {
	s.onClick(e.Modifier, s)
}

func (s *queueAlbumSong) MouseIn(e *desktop.MouseEvent) {
	s.hovered = true
	if s.isDragging() {
		s.showInsertMarker = true
		s.setDragMark(e.Position, true)
	}
	// TODO dieses ganze refresh-Geraffel ist hochgradig undurchsichtig
	// - normalerweise reicht canvas.Refresh(w) aber das triggert keinen Renderer.Refresh
	// - Renderer.Refresh soll wahrscheinlich aktualisieren (show/hide o.ä.) ohne layout/minsize-Änderung.
	// - Renderer.Refresh muss aber explizit gerufen werden und hängt damit von widget.Renderer() ab,
	//   was ja nun ganz und gar nicht geht.
	// - ausserdem is offenbar ein canvas.Refresh(renderer.widget) im Renderer gefährlich, wenn dessen widget embedded ist
	widget.Renderer(s).Refresh()
	canvas.Refresh(s)
}

func (s *queueAlbumSong) MouseMoved(e *desktop.MouseEvent) {
	if s.isDragging() {
		if s.setDragMark(e.Position, false) {
			widget.Renderer(s).Refresh()
			canvas.Refresh(s)
		}
	}
}

func (s *queueAlbumSong) MouseOut() {
	s.showInsertMarker = false
	s.hovered = false
	// TODO s.o.
	widget.Renderer(s).Refresh()
	canvas.Refresh(s)
}

func (s *queueAlbumSong) MouseUp(e *desktop.MouseEvent) {
}

func (s *queueAlbumSong) Resize(size fyne.Size) {
	s.resize(s, size)
}

func (s *queueAlbumSong) Tapped(_ *fyne.PointEvent) {
}

func (s *queueAlbumSong) TappedSecondary(e *fyne.PointEvent) {
	c := fyne.CurrentApp().Driver().CanvasForObject(s)

	items := []*fyne.MenuItem{
		fyne.NewMenuItem("Play now", func() { s.playNow(s.song) }),
		fyne.NewMenuItem("Remove", func() { s.removeFromQueue(s.song) }),
		fyne.NewMenuItem("Remove Others", func() { s.removeOthersFromQueue(s.song) }),
		fyne.NewMenuItem("Remove Before", func() { s.removeBeforeFromQueue(s.song) }),
		fyne.NewMenuItem("Remove After", func() { s.removeAfterFromQueue(s.song) }),
		fyne.NewMenuItem("Clear", s.clearQueue),
	}
	popUp := widget.NewPopUpMenu(fyne.NewMenu("", items...), c)

	// TODO provide absolute event position as part of the point event: e.AbsolutePosition
	pos := e.Position.Add(fyne.CurrentApp().Driver().AbsolutePositionForObject(s))
	popUp.Move(pos)
}

func (s *queueAlbumSong) Show() {
	s.show(s)
}

func (s *queueAlbumSong) setDragMark(p fyne.Position, force bool) bool {
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
