package ui

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
)

type queueAlbumLine struct {
	baseWidget
	bold               bool
	hovered            bool
	insertMarkerBottom bool
	lastTextRight      bool
	pad                int
	selected           bool
	showInsertMarker   bool
	texts              []string
}

func (e *queueAlbumLine) CreateRenderer() fyne.WidgetRenderer {
	texts := make([]fyne.CanvasObject, 0, len(e.texts))
	for _, t := range e.texts {
		text := canvas.NewText(t, theme.TextColor())
		if e.bold {
			text.TextStyle.Bold = true
		}
		texts = append(texts, text)
	}
	// TODO lines sind nicht horizontal (size x, 0)
	//sep := canvas.NewLine(theme.IconColor())
	//sep.StrokeWidth = 1
	sep := canvas.NewRectangle(theme.ButtonColor())
	insertMarker := canvas.NewRectangle(theme.IconColor())
	insertMarker.Hide()
	return &queueAlbumLineRenderer{
		baseRenderer:  baseRenderer{objects: append(texts, sep, insertMarker)},
		e:             e,
		insertMarker:  insertMarker,
		lastTextRight: e.lastTextRight,
		pad:           e.pad,
		sep:           sep,
		texts:         texts,
	}
}

func (e *queueAlbumLine) Hide() {
	e.hide(e)
}

func (e *queueAlbumLine) MinSize() fyne.Size {
	return e.minSize(e)
}

func (e *queueAlbumLine) Resize(size fyne.Size) {
	e.resize(e, size)
}

func (e *queueAlbumLine) Show() {
	e.show(e)
}

type queueAlbumLineRenderer struct {
	baseRenderer
	e             *queueAlbumLine
	insertMarker  fyne.CanvasObject
	lastTextRight bool
	minSize       fyne.Size
	pad           int
	sep           fyne.CanvasObject
	textHeight    int
	texts         []fyne.CanvasObject
}

func (r *queueAlbumLineRenderer) BackgroundColor() color.Color {
	if r.e.selected {
		return theme.PrimaryColor()
	}
	if r.e.hovered {
		return theme.HoverColor()
	}
	return color.Transparent
}

func (r *queueAlbumLineRenderer) Layout(size fyne.Size) {
	if (r.sep.Position() == fyne.Position{}) {
		offset := r.pad + theme.Padding()
		for _, t := range r.texts {
			t.Move(fyne.NewPos(offset, 0))
			offset += t.MinSize().Width + theme.Padding()
		}
		r.sep.Move(fyne.NewPos(0, r.textHeight))
	}
	if r.lastTextRight {
		t := r.texts[len(r.texts)-1]
		t.Move(fyne.NewPos(size.Width-t.MinSize().Width, 0))
	}
	r.sep.Resize(fyne.NewSize(size.Width, 1))
	r.insertMarker.Resize(fyne.NewSize(size.Width, 1))
}

func (r *queueAlbumLineRenderer) MinSize() fyne.Size {
	if (r.minSize == fyne.Size{}) {
		minWidth := r.pad + len(r.texts)*theme.Padding()
		for _, t := range r.texts {
			minWidth += t.MinSize().Width
			r.textHeight = fyne.Max(r.textHeight, t.MinSize().Height)
		}
		r.minSize = fyne.NewSize(minWidth, r.textHeight+1)
	}
	return r.minSize
}

func (r *queueAlbumLineRenderer) Refresh() {
	if r.e.showInsertMarker {
		if r.e.insertMarkerBottom {
			r.insertMarker.Move(fyne.NewPos(0, 20))
		} else {
			r.insertMarker.Move(fyne.NewPos(0, 0))
		}
		r.insertMarker.Show()
	} else {
		r.insertMarker.Hide()
	}
	// TODO siehe queueAlbumSong hovering (super lahm!)
	//canvas.Refresh(r.e)
}
