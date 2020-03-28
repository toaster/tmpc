package ui

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/widget"

	"github.com/toaster/tmpc/internal/mpd"
)

// Queue displays an MPD play queue.
type Queue struct {
	SongList
	contextMenu        *fyne.Menu
	currentPlayerState PlayerState
	currentSongIdx     int
	onClear            func()
	onPlay             func(*mpd.Song)
	onRemove           func([]*mpd.Song)
	pauseIndicator     fyne.CanvasObject
	playIndicator      fyne.CanvasObject
	stopIndicator      fyne.CanvasObject
}

// NewQueue returns a new queue.
func NewQueue(move func(*mpd.Song, int), onClear func(), onPlay func(*mpd.Song), onRemove func([]*mpd.Song)) *Queue {
	playIndicator := canvas.NewImageFromResource(rscPlayIndicator)
	playIndicator.Resize(fyne.NewSize(18, 18))
	pauseIndicator := canvas.NewImageFromResource(rscPauseIndicator)
	pauseIndicator.Resize(fyne.NewSize(18, 18))
	stopIndicator := canvas.NewImageFromResource(rscStopIndicator)
	stopIndicator.Resize(fyne.NewSize(18, 18))
	q := &Queue{
		SongList: SongList{
			box:          widget.NewVBox(),
			move:         move,
			supportsDrag: true,
		},
		onClear:        onClear,
		onPlay:         onPlay,
		onRemove:       onRemove,
		pauseIndicator: pauseIndicator,
		playIndicator:  playIndicator,
		stopIndicator:  stopIndicator,
	}
	q.ExtendBaseWidget(q)
	q.contextMenu = q.buildSongContextMenu()
	return q
}

// CurrentSongIndex returns the index in the queue of the currently played song.
func (q *Queue) CurrentSongIndex() int {
	return q.currentSongIdx
}

// SetCurrentSong sets the current song and the current player state for highlighting the
// corresponding entry.
func (q *Queue) SetCurrentSong(idx int, ps PlayerState) {
	q.currentSongIdx = idx
	q.currentPlayerState = ps

	offset := 0
	for _, a := range q.box.Children {
		album := a.(*songListAlbum)
		l := len(album.songs)
		album.removeSongIndicator()
		if offset <= idx && offset+l > idx {
			var indicator fyne.CanvasObject
			switch ps {
			case PlayerStatePlay:
				indicator = q.playIndicator
			case PlayerStatePause:
				indicator = q.pauseIndicator
			case PlayerStateStop:
				indicator = q.stopIndicator
			}

			album.setSongIndicator(idx-offset, indicator)
		}
		offset += l
	}
}

// Update replaces the current content of the queue with the given songs.
func (q *Queue) Update(songs []*mpd.Song) {
	q.SongList.Update(songs, q.onPlay, q.contextMenu)
}

func (q *Queue) buildSongContextMenu() *fyne.Menu {
	items := []*fyne.MenuItem{
		fyne.NewMenuItem("Play now", func() { q.onPlay(q.SelectedSongs()[0]) }),
		fyne.NewMenuItem("Remove", func() { q.onRemove(q.SelectedSongs()) }),
		fyne.NewMenuItem("Remove Others", func() { q.onRemove(q.NotSelectedSongs()) }),
		fyne.NewMenuItem("Remove Before", func() { q.removeBefore(q.SelectedSongs()[0]) }),
		fyne.NewMenuItem("Remove After", func() {
			songs := q.SelectedSongs()
			q.removeAfter(songs[len(songs)-1])
		}),
		fyne.NewMenuItem("Clear", q.onClear),
	}
	return fyne.NewMenu("", items...)
}

func (q *Queue) removeAfter(s *mpd.Song) {
	for i, song := range q.songs {
		if song == s {
			if i+1 < len(q.songs) {
				q.onRemove(q.songs[i+1:])
			}
			break
		}
	}
}

func (q *Queue) removeBefore(s *mpd.Song) {
	for i, song := range q.songs {
		if song == s {
			q.onRemove(q.songs[0:i])
			break
		}
	}
}
