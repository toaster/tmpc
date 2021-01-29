package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type listEntry struct {
	widget.BaseWidget
	hovered            bool
	insertMarkerBottom bool
	selected           bool
	showInsertMarker   bool
}

func (e *listEntry) createRenderer() listEntryRenderer {
	bg := canvas.NewRectangle(color.Transparent)
	sep := canvas.NewRectangle(theme.HoverColor())
	insertMarker := canvas.NewRectangle(theme.ForegroundColor())
	insertMarker.Hide()
	return listEntryRenderer{
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{bg, sep, insertMarker}},
		background:   bg,
		e:            e,
		insertMarker: insertMarker,
		sep:          sep,
	}
}

func (e *listEntry) Refresh() {
	// TODO: widget extension + WidgetRenderer + refreshing is still error-prone
	e.BaseWidget.Refresh()
	canvas.Refresh(e)
}

type listEntryRenderer struct {
	baseRenderer
	background   *canvas.Rectangle
	e            *listEntry
	insertMarker fyne.CanvasObject
	sep          fyne.CanvasObject
}

func (r *listEntryRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)
	r.sep.Move(fyne.NewPos(0, size.Height-1))
	r.sep.Resize(fyne.NewSize(size.Width, 1))
	r.insertMarker.Resize(fyne.NewSize(size.Width, 1))
}

func (r *listEntryRenderer) MinSize() fyne.Size {
	return fyne.NewSize(0, 1)
}

func (r *listEntryRenderer) Refresh() {
	if r.e.selected {
		r.background.FillColor = theme.PrimaryColor()
	} else if r.e.hovered {
		r.background.FillColor = theme.HoverColor()
	} else {
		r.background.FillColor = color.Transparent
	}

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
