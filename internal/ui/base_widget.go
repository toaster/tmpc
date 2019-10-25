package ui

import (
	"image/color"
	"log"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

type baseWidget struct {
	hidden bool
	pos    fyne.Position
	size   fyne.Size
}

// CreateRenderer is an internal method
func (b *baseWidget) CreateRenderer() fyne.WidgetRenderer {
	log.Fatalln("NOT IMPLEMENTED: CreateRenderer", b)
	return nil
}

func (b *baseWidget) Hide() {
	log.Println("NOT IMPLEMENTED: Hide")
}

func (b *baseWidget) hide(w fyne.Widget) {
	if b.hidden {
		return
	}
	b.hidden = true
	canvas.Refresh(w)
}

// MinSize returns the minimal size the widget needs to be painted.
func (b *baseWidget) MinSize() fyne.Size {
	log.Fatalln("NOT IMPLEMENTED: MinSize", b)
	return fyne.Size{}
}

func (b *baseWidget) minSize(w fyne.Widget) fyne.Size {
	//log.Println("MinSize")
	return widget.Renderer(w).MinSize()
}

func (b *baseWidget) Move(pos fyne.Position) {
	//log.Println("Move")
	b.pos = pos
	canvas.Refresh(b)
}

func (b *baseWidget) Position() fyne.Position {
	// called rapidly
	// log.Println("Position")
	return b.pos
}

// Resize resizes the widget to the given size.
func (b *baseWidget) Resize(size fyne.Size) {
	log.Fatalln("NOT IMPLEMENTED: Resize", b)
}

func (b *baseWidget) resize(w fyne.Widget, size fyne.Size) {
	//log.Println("Resize:", size)
	b.size = size
	widget.Renderer(w).Layout(size)
}

func (b *baseWidget) Show() {
	log.Fatalln("NOT IMPLEMENTED: Show", b)
}

func (b *baseWidget) show(w fyne.Widget) {
	if !b.hidden {
		return
	}
	b.hidden = false
	canvas.Refresh(w)
}

func (b *baseWidget) Size() fyne.Size {
	// called rapidly on hover
	// log.Println("Size:", b.size)
	return b.size
}

func (b *baseWidget) Visible() bool {
	//log.Println("Visible")
	return !b.hidden
}

type baseRenderer struct {
	objects []fyne.CanvasObject
}

func (b *baseRenderer) ApplyTheme() {
	log.Fatalln("NOT IMPLEMENTED: ApplyTheme", b)
	// TODO
}

func (b *baseRenderer) BackgroundColor() color.Color {
	//log.Println("BackgroundColor:", theme.BackgroundColor())
	return theme.BackgroundColor()
}

func (b *baseRenderer) Destroy() {
	//log.Println("Destroy")
}

func (b *baseRenderer) Objects() []fyne.CanvasObject {
	// called rapidly
	// log.Println("Objects")
	return b.objects
}

func (b *baseRenderer) Refresh() {
	log.Fatalln("NOT IMPLEMENTED: Refresh", b)
}
