package ui

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/theme"
)

type baseRenderer struct {
	objects []fyne.CanvasObject
}

func (b *baseRenderer) BackgroundColor() color.Color {
	// log.Println("BackgroundColor:", theme.BackgroundColor())
	return theme.BackgroundColor()
}

func (b *baseRenderer) Destroy() {
	// log.Println("Destroy")
}

func (b *baseRenderer) Objects() []fyne.CanvasObject {
	// called rapidly
	// log.Println("Objects")
	return b.objects
}
