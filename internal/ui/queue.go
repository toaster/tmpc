package ui

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/widget"
	"github.com/toaster/tmpc/internal/mpd"
	"github.com/toaster/tmpc/internal/repo"
)

// Queue displays an MPD play queue.
type Queue struct {
	baseWidget
	addedByLastMarkSlice map[*queueAlbumSong]bool
	albums               []*queueAlbum
	box                  *widget.Box
	coverRepo            *repo.CoverRepository
	currentPlayerState   PlayerState
	currentSongIdx       int
	dragBefore           *mpd.Song
	dragAfter            *mpd.Song
	dragTargetIsSelected bool
	markAnchor           *queueAlbumSong
	move                 func(*mpd.Song, int)
	onClear              func()
	onPlay               func(*mpd.Song)
	onRemove             func(*mpd.Song)
	selectionIsDragged   bool
	songs                []*mpd.Song
}

// NewQueue returns a new queue.
func NewQueue(move func(*mpd.Song, int), onClear func(), onPlay, onRemove func(*mpd.Song)) *Queue {
	return &Queue{
		box:      widget.NewVBox(),
		move:     move,
		onClear:  onClear,
		onPlay:   onPlay,
		onRemove: onRemove,
	}
}

// CreateRenderer is an internal method
func (q *Queue) CreateRenderer() fyne.WidgetRenderer {
	return q.box.CreateRenderer()
}

// IsEmpty returns whether this queue is empty or not.
func (q *Queue) IsEmpty() bool {
	return len(q.songs) == 0
}

// MinSize returns the minimal size the widget needs to be painted.
func (q *Queue) MinSize() fyne.Size {
	return q.minSize(q)
}

// Resize resizes the widget to the given size.
func (q *Queue) Resize(size fyne.Size) {
	q.resize(q, size)
}

// SetCurrentSong sets the current song and the current player state for highlighting the
// corresponding entry.
func (q *Queue) SetCurrentSong(idx int, ps PlayerState) {
	q.currentSongIdx = idx
	q.currentPlayerState = ps

	offset := 0
	for _, a := range q.box.Children {
		album := a.(*queueAlbum)
		l := len(album.songs)
		album.unsetCurrentSong()
		if offset <= idx && offset+l > idx {
			album.setCurrentSong(idx-offset, ps)
		}
		offset += l
	}
}

// Update replaces the current content of the queue with the given songs.
func (q *Queue) Update(songs []*mpd.Song) {
	oldAlbums := q.albums
	q.songs = songs
	q.box.Children = make([]fyne.CanvasObject, 0, len(q.songs))
	q.albums = make([]*queueAlbum, 0, len(q.songs))
	defer widget.Refresh(q)

	if len(q.songs) == 0 {
		return
	}

	queueSongs := map[int]*queueAlbumSong{}
	lastAlbum := q.songs[0].Album
	var albumStart int
	for i, song := range q.songs {
		if lastAlbum != song.Album {
			q.appendAlbum(i, q.songs[albumStart:i], queueSongs)
			albumStart = i
			lastAlbum = song.Album
		}
		if i == len(q.songs)-1 {
			q.appendAlbum(i+1, q.songs[albumStart:], queueSongs)
		}
	}
	for _, qa := range oldAlbums {
		for _, qs := range qa.songs {
			if qs.selected {
				queueSongs[qs.song.ID].selected = true
			}
			if q.markAnchor == qs {
				q.markAnchor = queueSongs[qs.song.ID]
			}
		}
	}
	q.recomputeAlbumSelections()
}

func (q *Queue) appendAlbum(i int, songs []*mpd.Song, queueSongs map[int]*queueAlbumSong) {
	album := newQueueAlbum(
		q.coverRepo.LoadCover(songs[0]),
		songs,
		q.markSong,
		q.markAlbum,
		q.dragSelection,
		func() bool { return q.selectionIsDragged },
		q.setDragMarkBefore,
		q.setDragMarkAfter,
		q.insertSelection,
		q.onPlay,
		q.onRemove,
		q.removeOthers,
		q.removeBefore,
		q.removeAfter,
		q.onClear,
	)
	q.box.Children = append(q.box.Children, album)
	q.albums = append(q.albums, album)
	for _, qs := range album.songs {
		queueSongs[qs.song.ID] = qs
	}
}

func (q *Queue) dragSelection() {
	q.selectionIsDragged = true
}

func (q *Queue) insertSelection() {
	q.selectionIsDragged = false
	q.moveSelection()
	q.dragBefore = nil
	q.dragAfter = nil
}

func (q *Queue) markAlbum(m desktop.Modifier, a *queueAlbum) {
	if m&desktop.ControlModifier != 0 {
		for _, qs := range a.songs {
			if !qs.selected {
				q.markSongWithoutAlbumRefresh(m, qs)
			}
		}
	} else if m&desktop.ShiftModifier != 0 {
		fromFront := false
	outerLoop:
		for _, qa := range q.albums {
			if qa == a {
				break
			}
			for _, qs := range qa.songs {
				if qs == q.markAnchor {
					fromFront = true
					break outerLoop
				}
			}
		}
		if fromFront {
			q.markSongWithoutAlbumRefresh(m, a.songs[len(a.songs)-1])
		} else {
			q.markSongWithoutAlbumRefresh(m, a.songs[0])
		}
	} else {
		m = 0
		for _, qs := range a.songs {
			q.markSongWithoutAlbumRefresh(m, qs)
			m = desktop.ControlModifier
		}
	}
	q.recomputeAlbumSelections()
}

func (q *Queue) markSongWithoutAlbumRefresh(m desktop.Modifier, s *queueAlbumSong) {
	var add, addSlice bool
	if m&desktop.ControlModifier != 0 {
		add = true
	} else if q.markAnchor != nil && m&desktop.ShiftModifier != 0 {
		addSlice = true
	}
	if addSlice {
		insideSlice := false
		for _, qa := range q.albums {
			for _, qs := range qa.songs {
				boundary := qs == q.markAnchor || qs == s
				if !insideSlice && boundary {
					insideSlice = true
					boundary = false
				}
				if insideSlice {
					q.addedByLastMarkSlice[qs] = q.addedByLastMarkSlice[qs] || !qs.selected
					qs.selected = true
					canvas.Refresh(qs)
				} else if q.addedByLastMarkSlice[qs] {
					qs.selected = false
					q.addedByLastMarkSlice[qs] = false
					canvas.Refresh(qs)
				}
				if insideSlice && boundary {
					insideSlice = false
				}
			}
		}
	} else {
		q.addedByLastMarkSlice = map[*queueAlbumSong]bool{}
		q.markAnchor = s
		if !add {
			for _, qa := range q.albums {
				for _, qs := range qa.songs {
					if qs != s && qs.selected {
						qs.selected = false
						canvas.Refresh(qs)
					}
				}
			}
		}
		s.selected = true
		canvas.Refresh(s)
	}
}

func (q *Queue) markSong(m desktop.Modifier, s *queueAlbumSong) {
	q.markSongWithoutAlbumRefresh(m, s)
	q.recomputeAlbumSelections()
}

func (q *Queue) moveSelection() {
	if q.dragTargetIsSelected {
		return
	}
	var target *mpd.Song
	targetIndex := 0
	if q.dragBefore != nil {
		target = q.dragBefore
	} else {
		target = q.dragAfter
		targetIndex++
	}
	beforeTargetCount := 0
	i := 0
	targetFound := false
	selection := []*mpd.Song{}
	for _, qa := range q.albums {
		for _, qs := range qa.songs {
			if qs.song == target {
				targetFound = true
			}
			if !targetFound {
				targetIndex++
			}
			if qs.selected {
				selection = append(selection, qs.song)
				if i < targetIndex {
					beforeTargetCount++
				}
			}
			i++
		}
	}
	for _, song := range selection[:beforeTargetCount] {
		// The song is removed by MPD before inserting and therefore we have to adjust the target index.
		q.move(song, targetIndex-1)
	}
	for i := len(selection) - 1; i >= beforeTargetCount; i-- {
		q.move(selection[i], targetIndex)
	}
}

func (q *Queue) recomputeAlbumSelections() {
	for _, qa := range q.albums {
		allSelected := true
		anySelected := false
		for _, qs := range qa.songs {
			if qs.selected {
				anySelected = true
			} else {
				allSelected = false
			}
		}
		qa.summary.selected = anySelected
		qa.header.selected = allSelected
		canvas.Refresh(qa.summary)
		canvas.Refresh(qa.header)
	}
}

func (q *Queue) removeAfter(s *mpd.Song) {
	remove := false
	for _, song := range q.songs {
		if remove {
			q.onRemove(song)
		}
		if song == s {
			remove = true
		}
	}
}

func (q *Queue) removeBefore(s *mpd.Song) {
	for _, song := range q.songs {
		if song == s {
			break
		}
		q.onRemove(song)
	}
}

func (q *Queue) removeOthers(s *mpd.Song) {
	for _, song := range q.songs {
		if song != s {
			q.onRemove(song)
		}
	}
}

func (q *Queue) setDragMarkAfter(s *mpd.Song, isSelected bool) {
	q.dragBefore = nil
	q.dragAfter = s
	q.dragTargetIsSelected = isSelected
}

func (q *Queue) setDragMarkBefore(s *mpd.Song, isSelected bool) {
	q.dragBefore = s
	q.dragAfter = nil
	q.dragTargetIsSelected = isSelected
}
