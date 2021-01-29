package ui

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

// StatusBar is the status bar to be displayed at the bottom of TMPC’s main window.
type StatusBar struct {
	widget.BaseWidget
	connected              bool
	errorCount             int
	onConnectClick         func()
	onErrorsClick          func()
	onPlaybackClick        func() bool
	onPlaybackConnectClick func()
	playing                bool
	playbackEnabled        bool
}

// NewStatusBar creates a new status bar.
func NewStatusBar(playbackEnabled bool, onConnectClick, onErrorsClick func(), onPlaybackClick func() bool, onPlaybackConnectClick func()) *StatusBar {
	b := &StatusBar{
		onConnectClick:         onConnectClick,
		onErrorsClick:          onErrorsClick,
		onPlaybackClick:        onPlaybackClick,
		onPlaybackConnectClick: onPlaybackConnectClick,
		playbackEnabled:        playbackEnabled,
	}
	b.ExtendBaseWidget(b)
	return b
}

// CreateRenderer is an internal function.
func (b *StatusBar) CreateRenderer() fyne.WidgetRenderer {
	separator := canvas.NewRectangle(theme.PlaceHolderColor())
	box := widget.NewHBox()
	r := &statusBarRenderer{
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{box, separator}},
		b:            b,
		box:          box,
		separator:    separator,
	}
	callback := func() {
		r.b.playbackEnabled = r.b.onPlaybackClick()
		r.Refresh()
	}
	r.playbackButton = newIconButton(HeadphonesIcon, callback)
	r.playbackButtonDisabled = newIconButton(theme.NewDisabledResource(HeadphonesIcon), callback)
	r.playbackButton.iconSize = fyne.NewSize(20, 20)
	r.playbackButtonDisabled.iconSize = fyne.NewSize(20, 20)
	r.playbackButton.pad = false
	r.playbackButtonDisabled.pad = false
	r.updateIcons()
	return r
}

// SetErrorCount sets the error count that is then displayed in the status bar.
func (b *StatusBar) SetErrorCount(c int) {
	b.errorCount = c
	b.Refresh()
}

// SetIsConnected sets the MPD connection state which is then displayed in the status bar.
func (b *StatusBar) SetIsConnected(c bool) {
	b.connected = c
	b.Refresh()
}

// SetIsPlaying sets the ShoutCast playback state which is then displayed in the status bar.
func (b *StatusBar) SetIsPlaying(p bool) {
	b.playing = p
	b.Refresh()
}

type statusBarRenderer struct {
	baseRenderer
	b                      *StatusBar
	box                    *widget.Box
	playbackButton         *iconButton
	playbackButtonDisabled *iconButton
	separator              fyne.CanvasObject
}

func (r *statusBarRenderer) Layout(size fyne.Size) {
	r.separator.Resize(fyne.NewSize(size.Width, 1))
	r.box.Resize(r.box.MinSize())
	r.box.Move(fyne.NewPos(size.Width-r.box.MinSize().Width-theme.Padding(), 1+theme.Padding()))
}

func (r *statusBarRenderer) MinSize() fyne.Size {
	return r.box.MinSize().Add(fyne.NewSize(theme.Padding()*2, 0)).Union(fyne.NewSize(0, 21+theme.Padding()))
}

func (r *statusBarRenderer) Refresh() {
	r.updateIcons()
	canvas.Refresh(r.b)
}

func (r *statusBarRenderer) updateIcons() {
	r.box.Children = r.box.Children[0:0]
	if r.b.playbackEnabled {
		r.box.Append(r.playbackButton)
	} else {
		r.box.Append(r.playbackButtonDisabled)
	}
	if r.b.playing {
		icon := canvas.NewImageFromResource(MusicIcon)
		icon.SetMinSize(fyne.NewSize(20, 20))
		r.box.Append(icon)
	} else {
		button := newIconButton(NoMusicIcon, r.b.onPlaybackConnectClick)
		button.iconSize = fyne.NewSize(20, 20)
		button.pad = false
		r.box.Append(button)
	}
	if r.b.connected {
		icon := canvas.NewImageFromResource(PluggedIcon)
		icon.SetMinSize(fyne.NewSize(20, 20))
		r.box.Append(icon)
	} else {
		button := newIconButton(UnpluggedIcon, r.b.onConnectClick)
		button.iconSize = fyne.NewSize(20, 20)
		button.pad = false
		r.box.Append(button)
	}
	if r.b.errorCount > 0 {
		button := newIconButton(ErrorIcon, r.b.onErrorsClick)
		button.iconSize = fyne.NewSize(20, 20)
		button.pad = false
		button.UpdateBadgeCount(r.b.errorCount)
		r.box.Append(button)
	}
}
