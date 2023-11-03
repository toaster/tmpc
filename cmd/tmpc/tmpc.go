package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/toaster/tmpc/internal/metadata"
	"github.com/toaster/tmpc/internal/metadata/archiveorg"
	"github.com/toaster/tmpc/internal/metadata/cache"
	"github.com/toaster/tmpc/internal/metadata/cascade"
	"github.com/toaster/tmpc/internal/metadata/discogs"
	"github.com/toaster/tmpc/internal/metadata/genius"
	"github.com/toaster/tmpc/internal/metadata/happidev"
	"github.com/toaster/tmpc/internal/mpd"
	"github.com/toaster/tmpc/internal/shoutcast"
	"github.com/toaster/tmpc/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type tmpc struct {
	coverRepo       metadata.CoverFetcher
	ctrls           *ui.PlayerControls
	errors          []string
	fyne            fyne.App
	info            *ui.SongInfo
	lyricsRepo      metadata.LyricsFetcher
	mpd             *mpd.Client
	playbackEnabled bool
	playlists       []*mpd.Playlist
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
	player := &tmpc{
		fyne: a,
		win:  a.NewWindow("Tilos Music Player Client"),
	}
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
	infoCont := container.NewScroll(player.info)
	player.playlistsList = ui.NewPlaylistList(player.handlePlayList, player.handleDeletePlaylist, player.win)
	player.queue = ui.NewQueue(player.moveSongInQueue, player.handleClearQueue, player.handleSongDetails, player.handlePlaySong, player.handleRemoveSongs, player.loadCover)
	player.search = ui.NewSearch(player.handleSearch, player.handleAddToQueue, player.handleInsertIntoQueue, player.handleReplaceQueue, player.handleAddToPlaylist, player.handleSongDetails, player.loadCover)

	mainContent := container.NewAppTabs(
		container.NewTabItemWithIcon("Queue", ui.QueueIcon, container.NewScroll(player.queue)),
		container.NewTabItemWithIcon("Playlists", ui.ListIcon, container.NewScroll(player.playlistsList)),
		container.NewTabItemWithIcon("Search", theme.SearchIcon(), player.search),
		container.NewTabItemWithIcon("Information", theme.InfoIcon(), infoCont),
	)
	mainContent.SetTabLocation(container.TabLocationLeading)

	player.statusBar = ui.NewStatusBar(player.playbackEnabled, player.connectMPD, player.showErrors, player.togglePlayback, player.startPlayback)

	mainGrid := ui.NewMainGrid(mainContent, player.ctrls, player.status, player.statusBar)
	player.win.SetContent(mainGrid)
	player.win.SetMaster()

	prefsItem := fyne.NewMenuItem("Preferences…", player.showSettings)
	// prefsItem.KeyEquivalent = ","
	appMenu := fyne.NewMenu("TMPC", prefsItem)

	updateItem := fyne.NewMenuItem("Update", player.updateDB)
	dbMenu := fyne.NewMenu("Database", updateItem)

	mainMenu := fyne.NewMainMenu(appMenu, dbMenu)
	player.win.SetMainMenu(mainMenu)

	const updateChannelBufSize = 100
	player.playlistsUpdate = make(chan bool, updateChannelBufSize)
	player.queueUpdate = make(chan bool, updateChannelBufSize)
	player.stateUpdate = make(chan bool, updateChannelBufSize)

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
			<-t.stateUpdate
			log.Println("UPDATE: state")
			t.updateState()
			log.Println("UPDATE done: state")
		}
	}()
	go func() {
		for {
			<-t.queueUpdate
			log.Println("UPDATE: queue")
			t.updateQueue()
			log.Println("UPDATE done: queue")
			t.stateUpdate <- true
		}
	}()
	go func() {
		for {
			<-t.playlistsUpdate
			log.Println("UPDATE: playlistsList")
			t.updatePlaylists()
			log.Println("UPDATE: playlistsList")
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
	t.applyTheme()
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
	t.lyricsRepo = cache.NewFSLyrics(
		cascade.NewLyrics([]metadata.LyricsFetcher{
			happidev.NewLyrics(t.fyne.Preferences().String("happiDevAPIKey")),
			genius.NewLyrics(t.fyne.Preferences().String("geniusAccessToken")),
		}),
	)
	t.coverRepo = cache.NewFSCover(
		cascade.NewCover([]metadata.CoverFetcher{
			archiveorg.NewCover(),
			discogs.NewCover(
				t.fyne.Preferences().String("discogsAPIKey"),
				t.fyne.Preferences().String("discogsAPISecret"),
			),
		}),
	)
	if connect {
		t.connectMPD()
	}
}

func (t *tmpc) applyTheme() {
	if t.fyne.Preferences().String("theme") == "Dark" {
		t.fyne.Settings().SetTheme(theme.DarkTheme())
	} else {
		t.fyne.Settings().SetTheme(theme.LightTheme())
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
			t.addError(fmt.Errorf("failed to add song to queue: %w", err))
		}
		if i == 0 {
			firstID = id
		}
	}
	if play {
		if err := t.mpd.PlayID(firstID); err != nil {
			t.addError(fmt.Errorf("failed to play added song: %w", err))
		}
	}
}

func (t *tmpc) connectMPD() {
	if err := t.mpd.Connect(); err != nil {
		t.addError(fmt.Errorf("failed to connect MPD client: %w", err))
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
				t.addError(fmt.Errorf("failed to add song to playlist: %w", err))
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
		t.addError(fmt.Errorf("failed to clear queue: %w", err))
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
		t.addError(fmt.Errorf("failed to play next title: %w", err))
		return false
	}
	return true
}

func (t *tmpc) handlePlayList(name string) {
	if !t.mpd.IsConnected() {
		return
	}

	if err := t.mpd.ClearQueue(); err != nil {
		t.addError(fmt.Errorf("failed to play playlist: %w", err))
		return
	}
	if err := t.mpd.PlaylistLoad(name); err != nil {
		t.addError(fmt.Errorf("failed to play playlist: %w", err))
		return
	}
	t.playCurrent()
}

func (t *tmpc) handlePlaySong(song *mpd.Song) {
	if !t.mpd.IsConnected() {
		return
	}

	if err := t.mpd.Play(song); err != nil {
		t.addError(fmt.Errorf("failed to play song: %w", err))
		return
	}
	t.startPlayback()
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
		t.addError(fmt.Errorf("failed to pause title: %w", err))
		return false
	}
	return true
}

func (t *tmpc) handlePrevTap() bool {
	if !t.mpd.IsConnected() {
		return false
	}

	if err := t.mpd.Prev(); err != nil {
		t.addError(fmt.Errorf("failed to play previous title: %w", err))
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
			t.addError(fmt.Errorf("failed to remove song: %w", err))
		}
	}
}

func (t *tmpc) handleReplaceQueue(songs []*mpd.Song, play bool) {
	t.handleClearQueue()
	t.handleAddToQueue(songs, play)
}

func (t *tmpc) handleSearch(key, value string) ([]*mpd.Song, bool) {
	const limit = 100
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
		t.addError(fmt.Errorf("failed to seek in title: %w", err))
	}
}

func (t *tmpc) handleSongDetails(song *mpd.Song) {
	details := container.NewVBox(
		container.NewHBox(
			widget.NewLabelWithStyle("Artist:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(song.Artist),
		),
		container.NewHBox(
			widget.NewLabelWithStyle("Title:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(song.Title),
		),
		container.NewHBox(
			widget.NewLabelWithStyle("Album:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(song.Album),
		),
		container.NewHBox(
			widget.NewLabelWithStyle("Album Artist:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(song.AlbumArtist),
		),
		container.NewHBox(
			widget.NewLabelWithStyle("File:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(song.File),
		),
	)
	d := dialog.NewCustom("Song Details", "Close", details, t.win)
	d.Show()
	d.Show()
}

func (t *tmpc) handleStopTap() bool {
	if !t.mpd.IsConnected() {
		return false
	}

	if err := t.mpd.Stop(); err != nil {
		t.addError(fmt.Errorf("failed to stop title: %w", err))
		return false
	}
	t.stopPlayback()
	return true
}

func (t *tmpc) loadCover(song *mpd.Song, coverDefault fyne.Resource, callback func(fyne.Resource)) {
	callback(coverDefault)
	if song == nil {
		return
	}
	go func() {
		cover, err := t.coverRepo.LoadCover(song)
		if err != nil {
			log.Println("failed loading cover:", err)
			return
		}
		callback(cover)
	}()
}

func (t *tmpc) makePrefsEntry(key string, placeholder string) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetText(t.fyne.Preferences().String(key))
	entry.SetPlaceHolder(placeholder)
	entry.OnChanged = func(s string) {
		t.fyne.Preferences().SetString(key, s)
	}
	return entry
}

func (t *tmpc) makePrefsSecretEntry(key string, placeholder string) *widget.Entry {
	entry := widget.NewPasswordEntry()
	entry.SetText(t.fyne.Preferences().String(key))
	entry.SetPlaceHolder(placeholder)
	entry.OnChanged = func(s string) {
		t.fyne.Preferences().SetString(key, s)
	}
	return entry
}

func (t *tmpc) moveSongInQueue(song *mpd.Song, index int) {
	if !t.mpd.IsConnected() {
		return
	}

	if err := t.mpd.Move(song, index); err != nil {
		t.addError(fmt.Errorf("failed to move song in queue: %w", err))
	}
}

func (t *tmpc) playCurrent() bool {
	if t.queue.IsEmpty() {
		return false
	}

	if err := t.mpd.PlayCurrent(); err != nil {
		t.addError(fmt.Errorf("failed to play current song: %w", err))
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
	c := t.win.Canvas()
	ws := c.Size()
	pos := fyne.NewPos(
		ws.Width-ts.Width-theme.Padding()*4,
		ws.Height-ts.Height-t.statusBar.Size().Height-theme.Padding()*3,
	)
	widget.ShowPopUpAtPosition(container.NewWithoutLayout(bg, txt), c, pos)
	t.errors = t.errors[0:0]
	t.statusBar.SetErrorCount(0)
}

func (t *tmpc) showSettings() {
	if t.settings != nil {
		t.settings.RequestFocus()
		return
	}
	themeSelector := widget.NewRadioGroup([]string{"Dark", "Light"}, func(s string) {
		t.fyne.Preferences().SetString("theme", s)
		t.applyTheme()
	})
	themeSelector.SetSelected(t.fyne.Preferences().String("theme"))
	themeSelector.Required = true
	themeSelector.Horizontal = true
	tmpDir, _ := cache.TmpDir()
	settingsContainer := container.New(
		layout.NewFormLayout(),
		widget.NewLabelWithStyle("Cache Directory", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		widget.NewLabel(tmpDir),
		widget.NewLabelWithStyle("MPD Server URL", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		t.makePrefsEntry("mpdURL", "mpd.example.com:6600"),
		widget.NewLabelWithStyle("MPD Server Password", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		t.makePrefsSecretEntry("mpdPass", "top secret MPD password"),
		widget.NewLabelWithStyle("Shoutcast Server URL", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		t.makePrefsEntry("shoutcastURL", "http://mpd.example.com:8000"),
		widget.NewLabelWithStyle("genius access token", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		t.makePrefsSecretEntry("geniusAccessToken", "top secret Genius access token"),
		widget.NewLabelWithStyle("happi.dev API key", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		t.makePrefsSecretEntry("happiDevAPIKey", "top secret happi.dev key"),
		widget.NewLabelWithStyle("Discogs API key", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		t.makePrefsEntry("discogsAPIKey", "key"),
		widget.NewLabelWithStyle("Discogs API secret", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		t.makePrefsSecretEntry("discogsAPISecret", "secret"),
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
		t.addError(fmt.Errorf("failed to start ShoutCast player: %w", err))
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

func (t *tmpc) updateDB() {
	if !t.mpd.IsConnected() {
		return
	}

	if err := t.mpd.Update(); err != nil {
		t.addError(fmt.Errorf("failed to update MPD database: %w", err))
	}
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
	t.loadCover(song, ui.AlbumIcon, t.status.UpdateCover)
	t.updateInfo(song)
}
