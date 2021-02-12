package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"github.com/toaster/tmpc/internal/mpd"
)

// Queue displays an MPD play queue.
type Queue struct {
	SongList
	contextMenu        *fyne.Menu
	currentPlayerState PlayerState
	currentSongIdx     int
	onClear            func()
	onDetails          func(*mpd.Song)
	onPlay             func(*mpd.Song)
	onRemove           func([]*mpd.Song)
	pauseIndicator     fyne.CanvasObject
	playIndicator      fyne.CanvasObject
	stopIndicator      fyne.CanvasObject
}

// NewQueue returns a new queue.
func NewQueue(move func(*mpd.Song, int), onClear func(), onDetails, onPlay func(*mpd.Song), onRemove func([]*mpd.Song), coverLoader func(*mpd.Song, fyne.Resource, func(fyne.Resource))) *Queue {
	playIndicator := canvas.NewImageFromResource(rscPlayIndicator)
	playIndicator.Resize(fyne.NewSize(18, 18))
	pauseIndicator := canvas.NewImageFromResource(rscPauseIndicator)
	pauseIndicator.Resize(fyne.NewSize(18, 18))
	stopIndicator := canvas.NewImageFromResource(rscStopIndicator)
	stopIndicator.Resize(fyne.NewSize(18, 18))
	q := &Queue{
		SongList: SongList{
			box:          container.NewVBox(),
			coverLoader:  coverLoader,
			move:         move,
			supportsDrag: true,
		},
		onClear:        onClear,
		onDetails:      onDetails,
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
	for _, a := range q.box.Objects {
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
		fyne.NewMenuItem("Play now", func() { q.onPlay(q.SongsSelected()[0]) }),
		fyne.NewMenuItem("Remove", func() { q.onRemove(q.SongsSelected()) }),
		fyne.NewMenuItem("Remove Others", func() { q.onRemove(q.SongsNotSelected()) }),
		fyne.NewMenuItem("Remove Before", func() { q.onRemove(q.SongsBeforeSelection()) }),
		fyne.NewMenuItem("Remove After", func() { q.onRemove(q.SongsAfterSelection()) }),
		fyne.NewMenuItem("Clear", q.onClear),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Details…", func() { q.onDetails(q.SongsSelected()[0]) }),
	}
	return fyne.NewMenu("", items...)
}
