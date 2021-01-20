package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/toaster/tmpc/internal/mpd"
)

// PlayerStatus is a widget displaying information on the currently playing song.
type PlayerStatus struct {
	widget.BaseWidget
	cover       fyne.Resource
	elapsed     int
	onSeek      func(int)
	progressBar *progressBar
	song        *mpd.Song
	ticker      *time.Ticker
}

// NewPlayerStatus returns a new PlayerStatus.
func NewPlayerStatus(onSeek func(int)) *PlayerStatus {
	s := &PlayerStatus{onSeek: onSeek}
	s.ExtendBaseWidget(s)
	return s
}

// CreateRenderer satisfies the fyne.Widget interface.
func (s *PlayerStatus) CreateRenderer() fyne.WidgetRenderer {
	aTitle := canvas.NewText("", theme.ForegroundColor())
	aCover := canvas.NewImageFromResource(nil)
	progress := canvas.NewText("0:00 / 0:00", theme.ForegroundColor())
	progress.TextSize = theme.TextSize() * 8 / 10
	s.progressBar = newProgressBar(0, 0, s.onSeek)
	title := canvas.NewText("", theme.ForegroundColor())
	title.TextStyle.Bold = true
	return &playerStatusRenderer{
		aCover:       aCover,
		aTitle:       aTitle,
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{aCover, aTitle, progress, s.progressBar, title}},
		progress:     progress,
		progressBar:  s.progressBar,
		s:            s,
		title:        title,
	}
}

// UpdateCover changes the cover.
func (s *PlayerStatus) UpdateCover(cover fyne.Resource) {
	s.cover = cover
	s.Refresh()
}

// UpdateSong updates the song, the progress bar, the playing state and the ticker.
func (s *PlayerStatus) UpdateSong(song *mpd.Song, elapsed int, playing bool) {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.song = song
	s.elapsed = elapsed
	s.Refresh()
	if playing {
		s.ticker = time.NewTicker(1 * time.Second)
		go func() {
			for range s.ticker.C {
				s.elapsed++
				s.Refresh()
			}
		}()
	}
}

type playerStatusRenderer struct {
	baseRenderer
	aCover      *canvas.Image
	aCoverSize  fyne.Size
	aTitle      *canvas.Text
	progress    *canvas.Text
	progressBar *progressBar
	s           *PlayerStatus
	song        *mpd.Song
	title       *canvas.Text
}

func (r *playerStatusRenderer) Layout(size fyne.Size) {
	r.title.Move(fyne.NewPos(r.aCoverSize.Width+theme.Padding(), 0))
	r.title.Resize(r.title.MinSize())
	r.aTitle.Move(fyne.NewPos(r.aCoverSize.Width+theme.Padding(), r.title.Size().Height))
	r.aTitle.Resize(r.aTitle.MinSize())
	r.progress.Move(fyne.NewPos(
		size.Width-r.progress.MinSize().Width-theme.Padding(),
		r.title.Size().Height+r.aTitle.Size().Height-r.progress.MinSize().Height,
	))
	r.progress.Resize(r.progress.MinSize())
	r.progressBar.Move(fyne.NewPos(r.aCoverSize.Width, r.title.Size().Height+r.aTitle.Size().Height))
	r.progressBar.Resize(fyne.NewSize(size.Width-r.aCoverSize.Width, r.progressBar.MinSize().Height))
	if r.aCover != nil {
		r.aCover.Resize(r.aCoverSize)
	}
}

func (r *playerStatusRenderer) MinSize() fyne.Size {
	titleMinSize := r.title.MinSize()
	aTitleMinSize := r.aTitle.MinSize()
	progressMinSize := r.progress.MinSize()
	progressBarMinSize := r.progressBar.MinSize()
	height := titleMinSize.Height + aTitleMinSize.Height + progressBarMinSize.Height
	r.aCoverSize = fyne.NewSize(height, height)
	return fyne.NewSize(
		r.aCoverSize.Width+fyne.Max(
			progressBarMinSize.Width,
			fyne.Max(titleMinSize.Width+theme.Padding()*2, aTitleMinSize.Width+progressMinSize.Width+theme.Padding()*3),
		),
		height,
	)
}

func (r *playerStatusRenderer) Refresh() {
	if r.title.Color != theme.ForegroundColor() {
		r.aTitle.Color = theme.ForegroundColor()
		r.progress.Color = theme.ForegroundColor()
		r.title.Color = theme.ForegroundColor()
	}
	if r.s.cover != r.aCover.Resource {
		r.aCover.Resource = r.s.cover
		canvas.Refresh(r.aCover)
	}
	if r.s.song != r.song {
		r.song = r.s.song
		if r.song == nil {
			r.progressBar.ReInit(0, 0, 0)
			r.progress.Text = "0:00 / 0:00"
			r.aTitle.Text = ""
			r.title.Text = ""
			canvas.Refresh(r.progress)
		} else {
			r.progressBar.ReInit(0, r.song.Time, 0)
			r.aTitle.Text = fmt.Sprintf("%s - %s (%d)", r.song.AlbumArtist, r.song.Album, r.song.Year)
			r.title.Text = r.song.DisplayTitle()
		}
		canvas.Refresh(r.aTitle)
		canvas.Refresh(r.title)
		canvas.Refresh(r.s)
	}
	if r.song != nil {
		r.progressBar.Update(r.s.elapsed)
		r.progress.Text = fmt.Sprintf("%d:%02d / %d:%02d", r.s.elapsed/60, r.s.elapsed%60, r.song.Time/60, r.song.Time%60)
		canvas.Refresh(r.progress)
	}
}
