package main

import (
	"flag"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"runtime/pprof"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/pkg/errors"
	"github.com/toaster/tmpc/internal/mpd"
	"github.com/toaster/tmpc/internal/repo"
	"github.com/toaster/tmpc/internal/shoutcast"
	"github.com/toaster/tmpc/internal/ui"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

type tmpc struct {
	coverRepo       *repo.CoverRepository
	ctrls           *ui.PlayerControls
	errors          []string
	fyne            fyne.App
	info            *ui.SongInfo
	lyricsRepo      *repo.LyricsRepository
	mpd             *mpd.Client
	playbackEnabled bool
	playlists       *widget.Box
	queue           *ui.Queue
	shoutcast       *shoutcast.Client
	status          *ui.PlayerStatus
	statusBar       *ui.StatusBar
	win             fyne.Window
}

func NewTMPC(t fyne.Theme, mpdUrl, mpdPass string, shoutcastUrl string) *tmpc {
	a := app.New()
	a.Settings().SetTheme(t)
	tmpc := &tmpc{
		fyne: a,
	}

	tmpc.mpd = mpd.NewClient(mpdUrl, mpdPass, tmpc.Update, tmpc.addError)
	tmpc.shoutcast = shoutcast.NewClient(shoutcastUrl, tmpc.addError)

	tmpc.ctrls = ui.NewPlayerControls(tmpc.handleNextTap, tmpc.handlePlayTap, tmpc.handlePauseTap,
		tmpc.handlePrevTap, tmpc.handleStopTap)
	tmpc.status = ui.NewPlayerStatus(tmpc.handleSeek)
	tmpc.info = ui.NewSongInfo()
	infoCont := widget.NewScrollContainer(tmpc.info)
	tmpc.playlists = widget.NewVBox()
	playlistsCont := widget.NewScrollContainer(tmpc.playlists)
	tmpc.queue = ui.NewQueue(tmpc.moveSongInQueue, tmpc.handleClearQueue, tmpc.handlePlaySong, tmpc.handleRemoveSong)
	queueCont := widget.NewScrollContainer(tmpc.queue)
	mainContent := widget.NewTabContainer(
		widget.NewTabItemWithIcon("Queue", ui.QueueIcon, queueCont),
		widget.NewTabItemWithIcon("Playlists", ui.ListIcon, playlistsCont),
		widget.NewTabItemWithIcon("Search", theme.SearchIcon(), widget.NewLabel("…")),
		widget.NewTabItemWithIcon("Information", theme.InfoIcon(), infoCont),
	)
	mainContent.SetTabLocation(widget.TabLocationLeading)

	tmpc.statusBar = ui.NewStatusBar(tmpc.playbackEnabled, tmpc.connectMPD, tmpc.showErrors, tmpc.togglePlayback, tmpc.startPlayback)

	mainGrid := ui.NewMainGrid(mainContent, tmpc.ctrls, tmpc.status, tmpc.statusBar)
	tmpc.win = tmpc.fyne.NewWindow("Tilos Music Player Client")
	tmpc.win.SetContent(mainGrid)

	return tmpc
}

func (t *tmpc) Run() {
	t.connectMPD()
	status, err := t.mpd.Status()
	if err == nil && status.State == mpd.StatePlay {
		t.startPlayback()
	}
	t.win.ShowAndRun()
}

func (t *tmpc) Update(subsystem string) {
	log.Println("UPDATE", subsystem)
	switch subsystem {
	case "player":
		t.UpdateState()
	case "playlist":
		t.UpdateQueue()
	case "stored_playlist":
		t.UpdatePlaylists()
	}
}

func (t *tmpc) UpdateInfo(song *mpd.Song) {
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

func (t *tmpc) UpdatePlaylists() {
	if !t.mpd.IsConnected() {
		return
	}

	pls, err := t.mpd.Playlists()
	if err != nil {
		log.Println("MPD update playlists error:", err)
	}

	t.playlists.Children = make([]fyne.CanvasObject, 0, len(pls))
	for _, pl := range pls {
		t.playlists.Children = append(t.playlists.Children, widget.NewLabel(pl.Name()))
	}

	widget.Refresh(t.playlists)
}

func (t *tmpc) UpdateQueue() {
	if !t.mpd.IsConnected() {
		return
	}

	songs, err := t.mpd.CurrentSongs()
	if err != nil {
		log.Println("MPD update queue error:", err)
	}

	t.queue.Update(songs)
	t.UpdateState()
}

func (t *tmpc) UpdateState() {
	if !t.mpd.IsConnected() {
		return
	}

	s, err := t.mpd.Status()
	if err != nil {
		log.Println("MPD update state error:", err)
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
	t.status.Update(t.coverRepo.LoadCover(song), song, s.Elapsed, s.State == mpd.StatePlay)
	t.UpdateInfo(song)
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

func (t *tmpc) addError(err error) {
	t.errors = append(t.errors, err.Error())
	t.statusBar.SetErrorCount(len(t.errors))
	t.statusBar.SetIsPlaying(t.shoutcast.IsPlaying())
	t.statusBar.SetIsConnected(t.mpd.IsConnected())
}

func (t *tmpc) connectMPD() {
	if err := t.mpd.Connect(); err != nil {
		t.addError(errors.Wrap(err, "failed to connect MPD client"))
	}
	t.statusBar.SetIsConnected(t.mpd.IsConnected())
	if t.mpd.IsConnected() {
		go func() {
			t.UpdateState()
			t.UpdateQueue()
			t.UpdatePlaylists()
		}()
		t.ctrls.Enable()
	} else {
		t.ctrls.Disable()
	}
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

func (t *tmpc) handleRemoveSong(song *mpd.Song) {
	if !t.mpd.IsConnected() {
		return
	}

	if err := t.mpd.RemoveFromQueue(song); err != nil {
		t.addError(errors.Wrap(err, "failed to remove song"))
		return
	}
	return
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

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	// TODO some sort of configuration
	t := NewTMPC(theme.LightTheme(),
		os.Getenv("TMPC_MPD_SERVER"),
		os.Getenv("TMPC_MPD_PASSWORD"),
		os.Getenv("TMPC_SHOUTCAST_URL"),
	)

	t.Run()
}
