package ui

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
)

type songListAlbumLine struct {
	listEntry
	bold          bool
	lastTextRight bool
	pad           int
	texts         []string
}

func newAlbumEntryLine(pad int, texts []string) *songListAlbumLine {
	l := &songListAlbumLine{
		listEntry: listEntry{insertMarkerBottom: true},
		pad:       pad,
		texts:     texts,
	}
	l.ExtendBaseWidget(l)
	return l
}

func newAlbumHeadLine(pad int, texts []string) *songListAlbumLine {
	l := newAlbumEntryLine(pad, texts)
	l.bold = true
	l.lastTextRight = true
	return l
}

func (l *songListAlbumLine) CreateRenderer() fyne.WidgetRenderer {
	texts := make([]fyne.CanvasObject, 0, len(l.texts))
	for _, t := range l.texts {
		text := canvas.NewText(t, theme.TextColor())
		if l.bold {
			text.TextStyle.Bold = true
		}
		texts = append(texts, text)
	}
	ler := l.listEntry.createRenderer()
	ler.objects = append(texts, ler.objects...)
	return &songListAlbumLineRenderer{
		listEntryRenderer: ler,
		l:                 l,
		lastTextRight:     l.lastTextRight,
		pad:               l.pad,
		texts:             texts,
	}
}

type songListAlbumLineRenderer struct {
	listEntryRenderer
	l             *songListAlbumLine
	lastTextRight bool
	minSize       fyne.Size
	pad           int
	textHeight    int
	texts         []fyne.CanvasObject
}

func (r *songListAlbumLineRenderer) Layout(size fyne.Size) {
	offset := r.pad + theme.Padding()
	for _, t := range r.texts {
		t.Move(fyne.NewPos(offset, 0))
		offset += t.MinSize().Width + theme.Padding()
	}
	if r.lastTextRight {
		t := r.texts[len(r.texts)-1]
		t.Move(fyne.NewPos(size.Width-t.MinSize().Width, 0))
	}
	r.listEntryRenderer.Layout(size)
}

func (r *songListAlbumLineRenderer) MinSize() fyne.Size {
	if (r.minSize == fyne.Size{}) {
		minWidth := r.pad + len(r.texts)*theme.Padding()
		for _, t := range r.texts {
			minWidth += t.MinSize().Width
			r.textHeight = fyne.Max(r.textHeight, t.MinSize().Height)
		}
		r.minSize = fyne.NewSize(minWidth, r.textHeight).Add(r.listEntryRenderer.MinSize())
	}
	return r.minSize
}
