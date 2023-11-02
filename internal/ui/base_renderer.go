package ui

import (
	"fyne.io/fyne/v2"
)

type baseRenderer struct {
	objects []fyne.CanvasObject
}

func (*baseRenderer) Destroy() {
	// log.Println("Destroy")
}

func (b *baseRenderer) Objects() []fyne.CanvasObject {
	// called rapidly
	// log.Println("Objects")
	return b.objects
}
