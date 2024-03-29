package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/toaster/tmpc/internal/mpd"
)

// SongList displays a list of MPD songs.
type SongList struct {
	widget.BaseWidget
	addedByLastMarkSlice map[*songListAlbumSong]bool
	albums               []*songListAlbum
	box                  *fyne.Container
	coverLoader          func(*mpd.Song, fyne.Resource, func(fyne.Resource))
	dragBefore           *mpd.Song
	dragAfter            *mpd.Song
	dragTargetIsSelected bool
	markAnchor           *songListAlbumSong
	move                 func(*mpd.Song, int)
	selectionIsDragged   bool
	songs                []*mpd.Song
	supportsDrag         bool
}

// NewSongList creates a new empty SongList.
func NewSongList(coverLoader func(*mpd.Song, fyne.Resource, func(fyne.Resource))) *SongList {
	l := &SongList{
		box:         container.NewVBox(),
		coverLoader: coverLoader,
		move:        func(*mpd.Song, int) {},
	}
	l.ExtendBaseWidget(l)
	return l
}

// CreateRenderer is an internal method
func (l *SongList) CreateRenderer() fyne.WidgetRenderer {
	return &containerRenderer{l.box}
}

// IsEmpty returns whether this song list is empty or not.
func (l *SongList) IsEmpty() bool {
	return len(l.songs) == 0
}

// SongsAfterSelection returns all songs after the current selection.
func (l *SongList) SongsAfterSelection() []*mpd.Song {
	selected := l.SongsSelected()
	if len(selected) == 0 {
		return nil
	}

	lastSelected := selected[len(selected)-1]
	for i, s := range l.songs {
		if lastSelected == s && i+1 < len(l.songs) {
			return l.songs[i+1:]
		}
	}
	return nil
}

// SongsBeforeSelection returns all songs before the current selection.
func (l *SongList) SongsBeforeSelection() []*mpd.Song {
	selected := l.SongsSelected()
	if len(selected) == 0 {
		return nil
	}

	for i, s := range l.songs {
		if selected[0] == s {
			return l.songs[0:i]
		}
	}
	return nil
}

// SongsNotSelected returns the songs that are not selected.
func (l *SongList) SongsNotSelected() []*mpd.Song {
	var songs []*mpd.Song
	for _, qa := range l.albums {
		for _, qs := range qa.songs {
			if !qs.selected {
				songs = append(songs, qs.song)
			}
		}
	}
	return songs
}

// SongsSelected returns the selected songs.
func (l *SongList) SongsSelected() []*mpd.Song {
	var songs []*mpd.Song
	for _, qa := range l.albums {
		for _, qs := range qa.songs {
			if qs.selected {
				songs = append(songs, qs.song)
			}
		}
	}
	return songs
}

// Update replaces the current content of the song list with the given songs.
func (l *SongList) Update(songs []*mpd.Song, onSongClick func(*mpd.Song), contextMenu *fyne.Menu) {
	oldAlbums := l.albums
	l.songs = songs
	l.box.Objects = make([]fyne.CanvasObject, 0, len(l.songs))
	l.albums = make([]*songListAlbum, 0, len(l.songs))
	defer l.Refresh()

	if len(l.songs) == 0 {
		return
	}

	songListSongs := map[int]*songListAlbumSong{}
	lastAlbum := l.songs[0].Album
	var albumStart int
	for i, song := range l.songs {
		if lastAlbum != song.Album {
			l.appendAlbum(l.songs[albumStart:i], songListSongs, onSongClick, contextMenu)
			albumStart = i
			lastAlbum = song.Album
		}
		if i == len(l.songs)-1 {
			l.appendAlbum(l.songs[albumStart:], songListSongs, onSongClick, contextMenu)
		}
	}
	for _, qa := range oldAlbums {
		for _, qs := range qa.songs {
			if qs.selected {
				if song := songListSongs[qs.song.ID]; song != nil {
					song.selected = true
					song.Refresh()
				}
			}
			if l.markAnchor == qs {
				l.markAnchor = songListSongs[qs.song.ID]
			}
		}
	}
	l.recomputeAlbumSelections()
}

func (l *SongList) appendAlbum(songs []*mpd.Song, songListSongs map[int]*songListAlbumSong, onSongClick func(*mpd.Song), contextMenu *fyne.Menu) {
	album := newSongListAlbum(
		songs,
		l.markSong,
		l.markAlbum,
		l.dragSelection,
		func() bool { return l.selectionIsDragged },
		l.setDragMarkBefore,
		l.setDragMarkAfter,
		l.insertSelection,
		onSongClick,
		contextMenu,
	)
	l.box.Objects = append(l.box.Objects, album)
	l.albums = append(l.albums, album)
	for _, qs := range album.songs {
		songListSongs[qs.song.ID] = qs
	}
	l.coverLoader(songs[0], AlbumIcon, album.UpdateCover)
}

func (l *SongList) dragSelection() {
	l.selectionIsDragged = l.supportsDrag
}

func (l *SongList) insertSelection() {
	if !l.supportsDrag {
		return
	}
	l.selectionIsDragged = false
	l.moveSelection()
	l.dragBefore = nil
	l.dragAfter = nil
}

func (l *SongList) markAlbum(m fyne.KeyModifier, a *songListAlbum) {
	if m&fyne.KeyModifierControl != 0 {
		for _, qs := range a.songs {
			if !qs.selected {
				l.markSongWithoutAlbumRefresh(m, qs)
			}
		}
	} else if m&fyne.KeyModifierShift != 0 {
		fromFront := false
	outerLoop:
		for _, qa := range l.albums {
			if qa == a {
				break
			}
			for _, qs := range qa.songs {
				if qs == l.markAnchor {
					fromFront = true
					break outerLoop
				}
			}
		}
		if fromFront {
			l.markSongWithoutAlbumRefresh(m, a.songs[len(a.songs)-1])
		} else {
			l.markSongWithoutAlbumRefresh(m, a.songs[0])
		}
	} else {
		m = 0
		for _, qs := range a.songs {
			l.markSongWithoutAlbumRefresh(m, qs)
			m = fyne.KeyModifierControl
		}
	}
	l.recomputeAlbumSelections()
}

func (l *SongList) markSongWithoutAlbumRefresh(m fyne.KeyModifier, s *songListAlbumSong) {
	var add, addSlice bool
	if m&fyne.KeyModifierControl != 0 {
		add = true
	} else if l.markAnchor != nil && m&fyne.KeyModifierShift != 0 {
		addSlice = true
	}
	if addSlice {
		insideSlice := false
		for _, qa := range l.albums {
			for _, qs := range qa.songs {
				boundary := qs == l.markAnchor || qs == s
				if !insideSlice && boundary {
					insideSlice = true
					boundary = false
				}
				if insideSlice {
					l.addedByLastMarkSlice[qs] = l.addedByLastMarkSlice[qs] || !qs.selected
					qs.selected = true
					qs.Refresh()
				} else if l.addedByLastMarkSlice[qs] {
					qs.selected = false
					l.addedByLastMarkSlice[qs] = false
					qs.Refresh()
				}
				if insideSlice && boundary {
					insideSlice = false
				}
			}
		}
	} else {
		l.addedByLastMarkSlice = map[*songListAlbumSong]bool{}
		l.markAnchor = s
		if !add {
			for _, qa := range l.albums {
				for _, qs := range qa.songs {
					if qs != s && qs.selected {
						qs.selected = false
						qs.Refresh()
					}
				}
			}
		}
		s.selected = true
		s.Refresh()
	}
}

func (l *SongList) markSong(m fyne.KeyModifier, s *songListAlbumSong) {
	l.markSongWithoutAlbumRefresh(m, s)
	l.recomputeAlbumSelections()
}

func (l *SongList) moveSelection() {
	if l.dragTargetIsSelected {
		return
	}
	var target *mpd.Song
	targetIndex := 0
	if l.dragBefore != nil {
		target = l.dragBefore
	} else {
		target = l.dragAfter
		targetIndex++
	}
	beforeTargetCount := 0
	idx := 0
	targetFound := false
	var selection []*mpd.Song
	for _, qa := range l.albums {
		for _, qs := range qa.songs {
			if qs.song == target {
				targetFound = true
			}
			if !targetFound {
				targetIndex++
			}
			if qs.selected {
				selection = append(selection, qs.song)
				if idx < targetIndex {
					beforeTargetCount++
				}
			}
			idx++
		}
	}
	for _, song := range selection[:beforeTargetCount] {
		// The song is removed by MPD before inserting and therefore we have to adjust the target index.
		l.move(song, targetIndex-1)
	}
	for i := len(selection) - 1; i >= beforeTargetCount; i-- {
		l.move(selection[i], targetIndex)
	}
}

func (l *SongList) recomputeAlbumSelections() {
	for _, qa := range l.albums {
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
		qa.summary.Refresh()
		qa.header.Refresh()
	}
}

func (l *SongList) setDragMarkAfter(s *mpd.Song, isSelected bool) {
	if !l.supportsDrag {
		return
	}
	l.dragBefore = nil
	l.dragAfter = s
	l.dragTargetIsSelected = isSelected
}

func (l *SongList) setDragMarkBefore(s *mpd.Song, isSelected bool) {
	if !l.supportsDrag {
		return
	}
	l.dragBefore = s
	l.dragAfter = nil
	l.dragTargetIsSelected = isSelected
}
