package ui

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"

	"github.com/toaster/tmpc/internal/mpd"
)

// SearchFn is a function to perform a search on the MPD server.
type SearchFn func(string, string) ([]*mpd.Song, bool)

// SongFn is a callback function that handles events on a slice of MPD songs.
type SongFn func([]*mpd.Song, bool)

// Search is the search page.
type Search struct {
	widget.BaseWidget
	box         *fyne.Container
	category    *widget.Select
	contextMenu *fyne.Menu
	doSearch    SearchFn
	input       *SubmitEntry
	results     *SongList
}

var _ fyne.Widget = (*Search)(nil)

// NewSearch creates a new search page.
func NewSearch(doSearch SearchFn, addToQueue, insertIntoQueue, replaceQueue, addToPlaylist SongFn, showDetails func(*mpd.Song)) *Search {
	s := &Search{
		doSearch: doSearch,
		results:  NewSongList(),
	}

	// TODO: localization
	s.category = &widget.Select{
		Options: []string{
			"Album",
			"Artist",
			"Genre",
			"Song",
		},
		PlaceHolder: "-",
		Selected:    "Song",
	}
	s.input = NewSubmitEntry(s.search)
	// TODO: auto submit?
	// 	timer := time.AfterFunc(1*time.Second, func() {
	// 		fmt.Println("search for", input.Text)
	// 	})
	// 	timer.Stop()
	// 	value.OnChanged = func(v string) {
	// 		fmt.Println("reset timer")
	// 		timer.Reset(500 * time.Millisecond)
	// 	}
	topLayout := layout.NewBorderLayout(nil, nil, s.category, nil)
	top := fyne.NewContainerWithLayout(topLayout, s.category, s.input)
	mainLayout := layout.NewBorderLayout(top, nil, nil, nil)
	s.box = fyne.NewContainerWithLayout(mainLayout, top, widget.NewScrollContainer(s.results))
	s.contextMenu = s.buildContextMenu(addToQueue, insertIntoQueue, replaceQueue, addToPlaylist, showDetails)

	s.ExtendBaseWidget(s)
	return s
}

// CreateRenderer is an internal method
func (s *Search) CreateRenderer() fyne.WidgetRenderer {
	return &searchRenderer{c: s.box}
}

func (s *Search) buildContextMenu(addToQueue, insertIntoQueue, replaceQueue, addToPlaylist SongFn, showDetails func(*mpd.Song)) *fyne.Menu {
	items := []*fyne.MenuItem{
		fyne.NewMenuItem("Append to Queue", func() { addToQueue(s.results.SelectedSongs(), false) }),
		fyne.NewMenuItem("Append to Queue and Play", func() {
			addToQueue(s.results.SelectedSongs(), true)
		}),
		fyne.NewMenuItem("Insert into Queue", func() { insertIntoQueue(s.results.SelectedSongs(), false) }),
		fyne.NewMenuItem("Insert into Queue and Play", func() {
			insertIntoQueue(s.results.SelectedSongs(), true)
		}),
		fyne.NewMenuItem("Replace Queue", func() { replaceQueue(s.results.SelectedSongs(), false) }),
		fyne.NewMenuItem("Replace Queue and Play", func() {
			replaceQueue(s.results.SelectedSongs(), true)
		}),
		fyne.NewMenuItem("Add To Playlist…", func() { addToPlaylist(s.results.SelectedSongs(), false) }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Details…", func() { showDetails(s.results.SelectedSongs()[0]) }),
	}
	return fyne.NewMenu("", items...)
}

func (s *Search) search(value string) {
	searchKey := "title"
	switch s.category.Selected {
	case "Album":
		searchKey = "album"
	case "Artist":
		searchKey = "artist"
	case "Genre":
		searchKey = "genre"
	}
	songs, _ := s.doSearch(searchKey, value)
	s.results.Update(songs, func(song *mpd.Song) {}, s.contextMenu)
	s.Refresh()
}

type searchRenderer struct {
	c *fyne.Container
}

var _ fyne.WidgetRenderer = (*searchRenderer)(nil)

func (r *searchRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

func (r *searchRenderer) Destroy() {
}

func (r *searchRenderer) Layout(size fyne.Size) {
	r.c.Resize(size)
}

func (r *searchRenderer) MinSize() fyne.Size {
	return r.c.MinSize()
}

func (r *searchRenderer) Objects() []fyne.CanvasObject {
	return r.c.Objects
}

func (r *searchRenderer) Refresh() {
	r.c.Refresh()
}
