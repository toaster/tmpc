package ui

import (
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/widget"
)

type SongInfo struct {
	baseWidget
	title  string
	lyrics []string
}

func NewSongInfo() *SongInfo {
	return &SongInfo{}
}

func (i *SongInfo) CreateRenderer() fyne.WidgetRenderer {
	title := widget.NewLabel(i.title)
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}
	text := widget.NewLabel(formatLyrics(i.lyrics))
	text.Alignment = fyne.TextAlignCenter
	return &songInfoRenderer{
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{title, text}},
		i:            i,
		text:         text,
		title:        title,
	}
}

func (i *SongInfo) Hide() {
	i.hide(i)
}

func (i *SongInfo) MinSize() fyne.Size {
	return i.minSize(i)
}

func (i *SongInfo) Resize(size fyne.Size) {
	i.resize(i, size)
}

func (i *SongInfo) Show() {
	i.show(i)
}

func (i *SongInfo) Update(title string, lyrics []string) {
	i.title = title
	i.lyrics = lyrics
	widget.Refresh(i)
}

type songInfoRenderer struct {
	baseRenderer
	i     *SongInfo
	title *widget.Label
	text  *widget.Label
}

func (r *songInfoRenderer) Layout(size fyne.Size) {
	r.title.Resize(fyne.NewSize(size.Width, r.title.MinSize().Height))
	r.text.Resize(fyne.NewSize(size.Width, r.text.MinSize().Height))
	r.text.Move(fyne.NewPos(0, 2*r.title.Size().Height))
}

func (r *songInfoRenderer) MinSize() fyne.Size {
	titleMinSize := r.title.MinSize()
	textMinSize := r.text.MinSize()
	return fyne.NewSize(fyne.Max(titleMinSize.Width, textMinSize.Width), titleMinSize.Height*2+textMinSize.Height)
}

func (r *songInfoRenderer) Refresh() {
	r.title.SetText(r.i.title)
	r.text.SetText(formatLyrics(r.i.lyrics))
	canvas.Refresh(r.i)
}

func formatLyrics(lyrics []string) string {
	return strings.TrimSpace(strings.Join(lyrics, "\n"))
}
