package ui

import (
	"fyne.io/fyne/v2"
)

func eventIsOn(e *fyne.PointEvent, o fyne.CanvasObject) bool {
	return eventIsInArea(e, o.Position(), o.Size())
}

func eventIsInArea(e *fyne.PointEvent, p fyne.Position, s fyne.Size) bool {
	return e.Position.X >= p.X && e.Position.X < s.Width+p.X &&
		e.Position.Y >= p.Y && e.Position.Y < s.Height+p.Y
}
