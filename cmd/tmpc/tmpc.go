package main

import (
	"log"
	"strings"

	"github.com/toaster/tmpc/internal/mpd"
	"github.com/toaster/tmpc/internal/repo"
	"github.com/toaster/tmpc/internal/shoutcast"
	"github.com/toaster/tmpc/internal/ui"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/pkg/errors"
)

type tmpc struct {
	coverRepo       *repo.CoverRepository
	ctrls           *ui.PlayerControls
	errors          []string
	fyne            fyne.App
	info            *ui.SongInfo
	lyricsRepo      *repo.LyricsRepository
	mpd             *mpd.Client
	playbackEnabled bool
	playlists       []mpd.Playlist
	playlistsList   *ui.PlaylistList
	playlistsUpdate chan bool
	stateUpdate     chan bool
	queueUpdate     chan bool
	queue           *ui.Queue
	search          *ui.Search
	settings        fyne.Window
	shoutcast       *shoutcast.Client
	status          *ui.PlayerStatus
	statusBar       *ui.StatusBar
	win             fyne.Window
}

func newTMPC() *tmpc {
	a := app.NewWithID("net.pruetz.tmpc")
	player := &tmpc{fyne: a}
	player.applySettings(false)

	player.ctrls = ui.NewPlayerControls(
		player.handleNextTap,
		player.handlePlayTap,
		player.handlePauseTap,
		player.handlePrevTap,
		player.handleStopTap,
	)
	player.status = ui.NewPlayerStatus(player.handleSeek)
	player.info = ui.NewSongInfo()
	infoCont := widget.NewScrollContainer(player.info)
	player.playlistsList = ui.NewPlaylistList(player.handlePlayList, player.handleDeletePlaylist)
	player.queue = ui.NewQueue(player.moveSongInQueue, player.handleClearQueue, player.handlePlaySong, player.handleRemoveSongs)
	player.search = ui.NewSearch(player.handleSearch, player.handleAddToQueue, player.handleInsertIntoQueue, player.handleReplaceQueue, player.handleAddToPlaylist)

	mainContent := widget.NewTabContainer(
		widget.NewTabItemWithIcon("Queue", ui.QueueIcon, widget.NewScrollContainer(player.queue)),
		widget.NewTabItemWithIcon("Playlists", ui.ListIcon, widget.NewScrollContainer(player.playlistsList)),
		widget.NewTabItemWithIcon("Search", theme.SearchIcon(), player.search),
		widget.NewTabItemWithIcon("Information", theme.InfoIcon(), infoCont),
	)
	mainContent.SetTabLocation(widget.TabLocationLeading)

	player.statusBar = ui.NewStatusBar(player.playbackEnabled, player.connectMPD, player.showErrors, player.togglePlayback, player.startPlayback)

	mainGrid := ui.NewMainGrid(mainContent, player.ctrls, player.status, player.statusBar)
	player.win = player.fyne.NewWindow("Tilos Music Player Client")
	player.win.SetContent(mainGrid)
	player.win.SetMaster()

	prefsItem := fyne.NewMenuItem("Preferences…", player.showSettings)
	prefsItem.PlaceInNativeMenu = true
	prefsItem.Separate = true
	prefsItem.KeyEquivalent = ","
	mainMenu := fyne.NewMainMenu(fyne.NewMenu("TMPC", prefsItem))
	player.win.SetMainMenu(mainMenu)

	player.playlistsUpdate = make(chan bool, 100)
	player.queueUpdate = make(chan bool, 100)
	player.stateUpdate = make(chan bool, 100)

	return player
}

// Run runs the music player app.
func (t *tmpc) Run() {
	t.connectMPD()
	status, err := t.mpd.Status()
	if err == nil && status.State == mpd.StatePlay {
		t.startPlayback()
	}
	go func() {
		for {
			select {
			case <-t.stateUpdate:
				log.Println("UPDATE: state")
				t.updateState()
				log.Println("UPDATE done: state")
			}
		}
	}()
	go func() {
		for {
			select {
			case <-t.queueUpdate:
				log.Println("UPDATE: queue")
				t.updateQueue()
				log.Println("UPDATE done: queue")
				t.stateUpdate <- true
			}
		}
	}()
	go func() {
		for {
			select {
			case <-t.playlistsUpdate:
				log.Println("UPDATE: playlistsList")
				t.updatePlaylists()
				log.Println("UPDATE: playlistsList")
			}
		}
	}()
	t.win.ShowAndRun()
}

// Update is a callback for the MPD watchdog and handles state updates.
func (t *tmpc) Update(subsystem string) {
	log.Println("UPDATE received:", subsystem)
	switch subsystem {
	case "player":
		t.stateUpdate <- true
	case "playlist":
		t.queueUpdate <- true
	case "stored_playlist":
		t.playlistsUpdate <- true
	}
	log.Println("UPDATE enqueued:", subsystem)
}

func (t *tmpc) applySettings(connect bool) {
	if t.fyne.Preferences().String("theme") == "Dark" {
		t.fyne.Settings().SetTheme(theme.DarkTheme())
	} else {
		t.fyne.Settings().SetTheme(theme.LightTheme())
	}
	if t.mpd != nil && t.mpd.IsConnected() {
		t.mpd.Disconnect()
	}
	t.mpd = mpd.NewClient(
		t.fyne.Preferences().String("mpdURL"),
		t.fyne.Preferences().String("mpdPass"),
		t.Update,
		t.addError,
	)
	t.shoutcast = shoutcast.NewClient(t.fyne.Preferences().String("shoutcastURL"), t.addError)
	if connect {
		t.connectMPD()
	}
}

func (t *tmpc) addError(err error) {
	t.errors = append(t.errors, err.Error())
	t.statusBar.SetErrorCount(len(t.errors))
	t.statusBar.SetIsPlaying(t.shoutcast.IsPlaying())
	t.statusBar.SetIsConnected(t.mpd.IsConnected())
}

func (t *tmpc) addSongsToQueue(songs []*mpd.Song, pos int, play bool) {
	if !t.mpd.IsConnected() {
		return
	}
	var firstID int
	for i, song := range songs {
		id, err := t.mpd.AddToQueue(song, pos)
		if pos >= 0 {
			pos++
		}
		if err != nil {
			t.addError(errors.Wrap(err, "failed to add song to queue"))
		}
		if i == 0 {
			firstID = id
		}
	}
	if play {
		if err := t.mpd.PlayID(firstID); err != nil {
			t.addError(errors.Wrap(err, "failed to play added song"))
		}
	}
}

func (t *tmpc) connectMPD() {
	if err := t.mpd.Connect(); err != nil {
		t.addError(errors.Wrap(err, "failed to connect MPD client"))
	}
	t.statusBar.SetIsConnected(t.mpd.IsConnected())
	if t.mpd.IsConnected() {
		go func() {
			t.stateUpdate <- true
			t.queueUpdate <- true
			t.playlistsUpdate <- true
		}()
		t.ctrls.Enable()
	} else {
		t.ctrls.Disable()
	}
}

func (t *tmpc) handleAddToPlaylist(songs []*mpd.Song, _ bool) {
	if !t.mpd.IsConnected() {
		return
	}
	var names []string
	for _, playlist := range t.playlists {
		names = append(names, playlist.Name())
	}
	content := widget.NewSelectEntry(names)
	content.SetPlaceHolder("Enter playlist name")
	dialog.ShowCustomConfirm("Input Playlist", "Ok", "Cancel", content, func(ok bool) {
		if !ok {
			return
		}
		for _, song := range songs {
			if err := t.mpd.AddToPlaylist(song, content.Text); err != nil {
				t.addError(errors.Wrap(err, "failed to add song to playlist"))
			}
		}
	}, t.win)
}

func (t *tmpc) handleAddToQueue(songs []*mpd.Song, play bool) {
	t.addSongsToQueue(songs, -1, play)
}

func (t *tmpc) handleClearQueue() {
	if !t.mpd.IsConnected() {
		return
	}
	if err := t.mpd.ClearQueue(); err != nil {
		t.addError(errors.Wrap(err, "failed to clear queue"))
		return
	}
}

func (t *tmpc) handleDeletePlaylist(name string) {
	if !t.mpd.IsConnected() {
		return
	}

	if err := t.mpd.PlaylistRemove(name); err != nil {
		t.addError(err)
	}
}

func (t *tmpc) handleInsertIntoQueue(songs []*mpd.Song, play bool) {
	t.addSongsToQueue(songs, t.queue.CurrentSongIndex()+1, play)
}

func (t *tmpc) handleNextTap() bool {
	if !t.mpd.IsConnected() {
		return false
	}

	if err := t.mpd.Next(); err != nil {
		t.addError(errors.Wrap(err, "failed to play next title"))
		return false
	}
	return true
}

func (t *tmpc) handlePlayList(name string) {
	if !t.mpd.IsConnected() {
		return
	}

	if err := t.mpd.ClearQueue(); err != nil {
		t.addError(errors.Wrap(err, "failed to play playlist"))
		return
	}
	if err := t.mpd.PlaylistLoad(name); err != nil {
		t.addError(errors.Wrap(err, "failed to play playlist"))
		return
	}
	t.playCurrent()
	return
}

func (t *tmpc) handlePlaySong(song *mpd.Song) {
	if !t.mpd.IsConnected() {
		return
	}

	if err := t.mpd.Play(song); err != nil {
		t.addError(errors.Wrap(err, "failed to play song"))
		return
	}
	t.startPlayback()
	return
}

func (t *tmpc) handlePlayTap() bool {
	if !t.mpd.IsConnected() {
		return false
	}
	return t.playCurrent()
}

func (t *tmpc) handlePauseTap() bool {
	if !t.mpd.IsConnected() {
		return false
	}

	if err := t.mpd.Pause(true); err != nil {
		t.addError(errors.Wrap(err, "failed to pause title"))
		return false
	}
	return true
}

func (t *tmpc) handlePrevTap() bool {
	if !t.mpd.IsConnected() {
		return false
	}

	if err := t.mpd.Prev(); err != nil {
		t.addError(errors.Wrap(err, "failed to play previous title"))
		return false
	}
	return false
}

func (t *tmpc) handleRemoveSongs(songs []*mpd.Song) {
	if !t.mpd.IsConnected() {
		return
	}

	for _, song := range songs {
		if err := t.mpd.RemoveFromQueue(song); err != nil {
			t.addError(errors.Wrap(err, "failed to remove song"))
		}
	}
	return
}

func (t *tmpc) handleReplaceQueue(songs []*mpd.Song, play bool) {
	t.handleClearQueue()
	t.handleAddToQueue(songs, play)
}

func (t *tmpc) handleSearch(key, value string) ([]*mpd.Song, bool) {
	limit := 100
	songs, err := t.mpd.Search(key, value, limit)
	if err != nil {
		t.addError(err)
		return nil, false
	}
	return songs, len(songs) == limit
}

func (t *tmpc) handleSeek(time int) {
	if !t.mpd.IsConnected() {
		return
	}

	if err := t.mpd.Seek(time); err != nil {
		t.addError(errors.Wrap(err, "failed to seek in title"))
	}
}

func (t *tmpc) handleStopTap() bool {
	if !t.mpd.IsConnected() {
		return false
	}

	if err := t.mpd.Stop(); err != nil {
		t.addError(errors.Wrap(err, "failed to stop title"))
		return false
	}
	t.stopPlayback()
	return true
}

func (t *tmpc) moveSongInQueue(song *mpd.Song, index int) {
	if !t.mpd.IsConnected() {
		return
	}

	if err := t.mpd.Move(song, index); err != nil {
		t.addError(errors.Wrap(err, "failed to move song in queue"))
	}
}

func (t *tmpc) playCurrent() bool {
	if t.queue.IsEmpty() {
		return false
	}

	if err := t.mpd.PlayCurrent(); err != nil {
		t.addError(errors.Wrap(err, "failed to play current song"))
		return false
	}
	t.startPlayback()
	return true
}

func (t *tmpc) showErrors() {
	txt := widget.NewLabel(strings.ReplaceAll("• "+strings.Join(t.errors, "\n• "), ": ", ":\n   "))
	bg := canvas.NewRectangle(theme.BackgroundColor())
	ts := txt.MinSize()
	bg.SetMinSize(ts)
	c := fyne.NewContainer(bg, txt)
	errorInfo := widget.NewPopUp(c, t.win.Canvas())
	ws := t.win.Canvas().Size()
	errorInfo.Move(fyne.NewPos(
		ws.Width-ts.Width-theme.Padding()*4,
		ws.Height-ts.Height-t.statusBar.Size().Height-theme.Padding()*3,
	))
	t.errors = t.errors[0:0]
	t.statusBar.SetErrorCount(0)
}

func (t *tmpc) showSettings() {
	if t.settings != nil {
		t.settings.RequestFocus()
		return
	}
	urlEntry := widget.NewEntry()
	urlEntry.SetText(t.fyne.Preferences().String("mpdURL"))
	urlEntry.SetPlaceHolder("mpd.example.com:6600")
	urlEntry.OnChanged = func(s string) {
		t.fyne.Preferences().SetString("mpdURL", s)
	}
	passEntry := widget.NewPasswordEntry()
	passEntry.SetText(t.fyne.Preferences().String("mpdPass"))
	passEntry.SetPlaceHolder("top secret")
	passEntry.OnChanged = func(s string) {
		t.fyne.Preferences().SetString("mpdPass", s)
	}
	shoutcastURLEntry := widget.NewEntry()
	shoutcastURLEntry.SetText(t.fyne.Preferences().String("shoutcastURL"))
	shoutcastURLEntry.SetPlaceHolder("http://mpd.example.com:8000")
	shoutcastURLEntry.OnChanged = func(s string) {
		t.fyne.Preferences().SetString("shoutcastURL", s)
	}
	themeSelector := widget.NewRadio([]string{"Dark", "Light"}, func(s string) {
		t.fyne.Preferences().SetString("theme", s)
	})
	themeSelector.SetSelected(t.fyne.Preferences().String("theme"))
	themeSelector.SetMandatory(true)
	themeSelector.Horizontal = true
	settingsContainer := fyne.NewContainerWithLayout(
		layout.NewFormLayout(),
		widget.NewLabelWithStyle("MPD Server URL", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		urlEntry,
		widget.NewLabelWithStyle("MPD Server Password", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		passEntry,
		widget.NewLabelWithStyle("Shoutcast Server URL", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		shoutcastURLEntry,
		widget.NewLabelWithStyle("Theme", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		themeSelector,
	)
	t.settings = t.fyne.NewWindow("Settings")
	t.settings = fyne.CurrentApp().NewWindow("Settings")
	t.settings.SetContent(settingsContainer)
	t.settings.CenterOnScreen()
	t.settings.Show()
	t.settings.SetOnClosed(func() {
		t.applySettings(true)
		t.settings = nil
	})
}

func (t *tmpc) startPlayback() {
	if !t.playbackEnabled {
		return
	}
	if err := t.shoutcast.Play(); err != nil {
		t.addError(errors.Wrap(err, "Failed to start ShoutCast player"))
	}
	t.statusBar.SetIsPlaying(t.shoutcast.IsPlaying())
}

func (t *tmpc) stopPlayback() {
	t.shoutcast.Stop()
	t.statusBar.SetIsPlaying(t.shoutcast.IsPlaying())
}

func (t *tmpc) togglePlayback() bool {
	if t.playbackEnabled {
		t.playbackEnabled = false
		t.stopPlayback()
	} else {
		t.playbackEnabled = true
		t.startPlayback()
	}
	return t.playbackEnabled
}

func (t *tmpc) updateInfo(song *mpd.Song) {
	if song == nil {
		t.info.Update("", []string{})
		return
	}
	lines, err := t.lyricsRepo.FetchLyrics(song)
	if err != nil {
		log.Println("Fetch lyrics error:", err)
		t.info.Update(song.Title, []string{"No lyrics"})
		return
	}

	t.info.Update(song.Title, lines)
}

func (t *tmpc) updatePlaylists() {
	if !t.mpd.IsConnected() {
		return
	}

	var err error
	t.playlists, err = t.mpd.Playlists()
	if err != nil {
		log.Println("MPD update playlistsList error:", err)
	}

	t.playlistsList.Update(t.playlists)
}

func (t *tmpc) updateQueue() {
	if !t.mpd.IsConnected() {
		return
	}

	songs, err := t.mpd.CurrentSongs()
	if err != nil {
		log.Println("MPD update queue error:", err)
	}

	t.queue.Update(songs)
}

func (t *tmpc) updateState() {
	if !t.mpd.IsConnected() {
		return
	}

	s, err := t.mpd.Status()
	if err != nil {
		log.Println("MPD update state error:", err)
		return
	}

	var cs ui.PlayerState
	switch s.State {
	case mpd.StatePlay:
		cs = ui.PlayerStatePlay
	case mpd.StatePause:
		cs = ui.PlayerStatePause
	default:
		cs = ui.PlayerStateStop
	}
	t.ctrls.SetState(cs)
	t.queue.SetCurrentSong(s.SongIdx, cs)

	// TODO nur neu laden wenn nötig -> im tmpc cachen -> identifizierbar über die playlistID
	songs, err := t.mpd.CurrentSongs()
	if err != nil {
		log.Println("MPD update state error:", err)
	}
	var song *mpd.Song
	if len(songs) > 0 {
		song = songs[s.SongIdx]
	}
	t.status.UpdateSong(song, s.Elapsed, s.State == mpd.StatePlay)
	t.coverRepo.LoadCover(song, ui.AlbumIcon, t.status.UpdateCover)
	t.updateInfo(song)
}
