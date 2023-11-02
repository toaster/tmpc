package ui

import "fyne.io/fyne/v2"

var _ fyne.WidgetRenderer = (*containerRenderer)(nil)

type containerRenderer struct {
	b *fyne.Container
}

func (*containerRenderer) Destroy() {
}

func (s *containerRenderer) Objects() []fyne.CanvasObject {
	return s.b.Objects
}

func (s *containerRenderer) Layout(size fyne.Size) {
	s.b.Layout.Layout(s.b.Objects, size)
}

func (s *containerRenderer) MinSize() fyne.Size {
	return s.b.Layout.MinSize(s.b.Objects)
}

func (*containerRenderer) Refresh() {
}
