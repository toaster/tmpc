package ui

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/widget"
)

var _ fyne.Tappable = (*songListAlbumCover)(nil)

type songListAlbumCover struct {
	widget.BaseWidget
	image   *canvas.Image
	onClick func()
}

func newSongListAlbumCover(img *canvas.Image, onClick func()) *songListAlbumCover {
	c := &songListAlbumCover{image: img, onClick: onClick}
	c.ExtendBaseWidget(c)
	return c
}

func (c *songListAlbumCover) CreateRenderer() fyne.WidgetRenderer {
	return &songListAlbumCoverRenderer{
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{c.image}},
		c:            c,
		image:        c.image,
	}
}

func (c *songListAlbumCover) MinSize() fyne.Size {
	return c.image.MinSize()
}

func (c *songListAlbumCover) Tapped(e *fyne.PointEvent) {
	c.onClick()
}

func (c *songListAlbumCover) TappedSecondary(*fyne.PointEvent) {}

// Update changes the image.
func (c *songListAlbumCover) Update(image fyne.Resource) {
	c.image.Resource = image
	canvas.Refresh(c.image)
}

type songListAlbumCoverRenderer struct {
	baseRenderer
	c     *songListAlbumCover
	image fyne.CanvasObject
}

func (r *songListAlbumCoverRenderer) Layout(size fyne.Size) {
	r.image.Resize(size)
}

func (r *songListAlbumCoverRenderer) MinSize() fyne.Size {
	return r.image.MinSize()
}

func (r *songListAlbumCoverRenderer) Refresh() {
	r.Layout(r.c.Size())
	canvas.Refresh(r.c)
}
