package ui

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

type listEntry struct {
	widget.BaseWidget
	hovered            bool
	insertMarkerBottom bool
	selected           bool
	showInsertMarker   bool
}

func (e *listEntry) createRenderer() listEntryRenderer {
	// TODO lines sind nicht horizontal (size x, 0)
	// sep := canvas.NewLine(theme.IconColor()
	// sep.StrokeWidth = 1
	sep := canvas.NewRectangle(theme.ButtonColor())
	insertMarker := canvas.NewRectangle(theme.IconColor())
	insertMarker.Hide()
	return listEntryRenderer{
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{sep, insertMarker}},
		e:            e,
		insertMarker: insertMarker,
		sep:          sep,
	}
}

type listEntryRenderer struct {
	baseRenderer
	e            *listEntry
	insertMarker fyne.CanvasObject
	sep          fyne.CanvasObject
}

func (r *listEntryRenderer) BackgroundColor() color.Color {
	if r.e.selected {
		return theme.PrimaryColor()
	}
	if r.e.hovered {
		return theme.HoverColor()
	}
	return color.Transparent
}

func (r *listEntryRenderer) Layout(size fyne.Size) {
	r.sep.Move(fyne.NewPos(0, size.Height-1))
	r.sep.Resize(fyne.NewSize(size.Width, 1))
	r.insertMarker.Resize(fyne.NewSize(size.Width, 1))
}

func (r *listEntryRenderer) MinSize() fyne.Size {
	return fyne.NewSize(0, 1)
}

func (r *listEntryRenderer) Refresh() {
	if r.e.showInsertMarker {
		if r.e.insertMarkerBottom {
			r.insertMarker.Move(fyne.NewPos(0, r.e.Size().Height-1))
		} else {
			r.insertMarker.Move(fyne.NewPos(0, 0))
		}
		r.insertMarker.Show()
	} else {
		r.insertMarker.Hide()
	}
}
