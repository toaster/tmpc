package ui

import (
	"fyne.io/fyne"
)

var _ fyne.Tappable = (*queueAlbumCover)(nil)

type queueAlbumCover struct {
	baseWidget
	image   fyne.CanvasObject
	onClick func()
}

func (c *queueAlbumCover) CreateRenderer() fyne.WidgetRenderer {
	return &queueAlbumCoverRenderer{
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{c.image}},
		image:        c.image,
	}
}

func (c *queueAlbumCover) MinSize() fyne.Size {
	return c.image.MinSize()
}

func (c *queueAlbumCover) Resize(size fyne.Size) {
	c.resize(c, size)
}

func (c *queueAlbumCover) Tapped(e *fyne.PointEvent) {
	c.onClick()
}

func (c *queueAlbumCover) TappedSecondary(*fyne.PointEvent) {}

type queueAlbumCoverRenderer struct {
	baseRenderer
	image fyne.CanvasObject
}

func (r *queueAlbumCoverRenderer) Layout(size fyne.Size) {
	r.image.Resize(size)
}

func (r *queueAlbumCoverRenderer) MinSize() fyne.Size {
	return r.image.MinSize()
}
