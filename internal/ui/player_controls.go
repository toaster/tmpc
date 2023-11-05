package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// PlayerState represents the play state of the music player.
type PlayerState int

// Valid values for PlayerState.
const (
	PlayerStateStop PlayerState = iota
	PlayerStatePlay
	PlayerStatePause
)

// PlayerControls is a music player controls widget.
type PlayerControls struct {
	widget.BaseWidget
	disabled bool
	onNext   func() bool
	onPause  func() bool
	onPlay   func() bool
	onPrev   func() bool
	onStop   func() bool
	pauseBtn *iconButton
	playBtn  *iconButton
	nextBtn  *iconButton
	prevBtn  *iconButton
	stopBtn  *iconButton
}

// NewPlayerControls returns a new PlayerControls.
func NewPlayerControls(onNext, onPlay, onPause, onPrev, onStop func() bool) *PlayerControls {
	p := &PlayerControls{
		onNext:  onNext,
		onPause: onPause,
		onPlay:  onPlay,
		onPrev:  onPrev,
		onStop:  onStop,
	}
	p.nextBtn = newIconButton(iconNext, p.handleNext)
	p.pauseBtn = newIconButton(iconPause, p.handlePause)
	p.playBtn = newIconButton(iconPlay, p.handlePlay)
	p.prevBtn = newIconButton(iconPrev, p.handlePrev)
	p.stopBtn = newIconButton(iconStop, p.handleStop)
	p.ExtendBaseWidget(p)
	return p
}

// CreateRenderer satisfies the fyne.Widget interface.
func (p *PlayerControls) CreateRenderer() fyne.WidgetRenderer {
	return &playerControlsRenderer{
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{
			p.nextBtn, p.playBtn, p.pauseBtn, p.prevBtn, p.stopBtn}},
		c:        p,
		nextBtn:  p.nextBtn,
		playBtn:  p.playBtn,
		pauseBtn: p.pauseBtn,
		prevBtn:  p.prevBtn,
		stopBtn:  p.stopBtn,
	}
}

// Disable satisfies the fyne.Disableable interface.
func (p *PlayerControls) Disable() {
	p.disabled = true
	p.nextBtn.Disable()
	p.pauseBtn.Disable()
	p.playBtn.Disable()
	p.prevBtn.Disable()
	p.stopBtn.Disable()
}

// Enable satisfies the fyne.Disableable interface.
func (p *PlayerControls) Enable() {
	p.disabled = false
	p.pauseBtn.Enable()
	p.playBtn.Enable()
}

// SetState updates the player state displayed by the controls.
func (p *PlayerControls) SetState(state PlayerState) {
	playOrPauseHovered := p.pauseBtn.hovered || p.playBtn.hovered
	switch state {
	case PlayerStatePlay:
		p.pauseBtn.hovered = playOrPauseHovered
		p.pauseBtn.Show()
		p.playBtn.Hide()
		p.nextBtn.Enable()
		p.prevBtn.Enable()
		p.stopBtn.Enable()
	case PlayerStatePause:
		p.pauseBtn.Hide()
		p.playBtn.hovered = playOrPauseHovered
		p.playBtn.Show()
		p.nextBtn.Enable()
		p.prevBtn.Enable()
		p.stopBtn.Enable()
	default:
		p.pauseBtn.Hide()
		p.playBtn.hovered = playOrPauseHovered
		p.playBtn.Show()
		p.nextBtn.Disable()
		p.prevBtn.Disable()
		p.stopBtn.Disable()
	}
}

func (p *PlayerControls) handleNext() {
	if p.onNext() {
		p.SetState(PlayerStatePlay)
	}
}

func (p *PlayerControls) handlePause() {
	if p.onPause() {
		p.SetState(PlayerStatePause)
	}
}

func (p *PlayerControls) handlePlay() {
	if p.onPlay() {
		p.SetState(PlayerStatePlay)
	}
}

func (p *PlayerControls) handlePrev() {
	if p.onPrev() {
		p.SetState(PlayerStatePlay)
	}
}

func (p *PlayerControls) handleStop() {
	if p.onStop() {
		p.SetState(PlayerStateStop)
	}
}

type playerControlsRenderer struct {
	baseRenderer
	c        *PlayerControls
	nextBtn  fyne.CanvasObject
	playBtn  fyne.CanvasObject
	pauseBtn fyne.CanvasObject
	prevBtn  fyne.CanvasObject
	ppSize   fyne.Size
	stopBtn  fyne.CanvasObject
}

func (p *playerControlsRenderer) Layout(_ fyne.Size) {
	if (p.playBtn.Size() == fyne.Size{}) {
		p.prevBtn.Resize(p.prevBtn.MinSize())
		p.playBtn.Resize(p.ppSize)
		ppx := p.prevBtn.Size().Width
		p.playBtn.Move(fyne.NewPos(ppx, 0))
		p.pauseBtn.Resize(p.ppSize)
		p.pauseBtn.Move(fyne.NewPos(ppx, 0))
		p.stopBtn.Resize(p.stopBtn.MinSize())
		sx := ppx + p.ppSize.Width
		p.stopBtn.Move(fyne.NewPos(sx, 0))
		p.nextBtn.Resize(p.nextBtn.MinSize())
		nx := sx + p.stopBtn.Size().Width
		p.nextBtn.Move(fyne.NewPos(nx, 0))
	}
}

func (p *playerControlsRenderer) MinSize() fyne.Size {
	p.ppSize = p.playBtn.MinSize().Max(p.pauseBtn.MinSize())
	ns := p.nextBtn.MinSize()
	ps := p.prevBtn.MinSize()
	ss := p.stopBtn.MinSize()
	return fyne.NewSize(p.ppSize.Width+ns.Width+ps.Width+ss.Width,
		fyne.Max(fyne.Max(p.ppSize.Height, ns.Height), fyne.Max(ps.Height, ss.Height)))
}

func (p *playerControlsRenderer) Refresh() {
	p.Layout(p.c.Size())
	canvas.Refresh(p.c)
}
