package ui

import (
	"fyne.io/fyne"
	"fyne.io/fyne/theme"
)

// TODO
// - scrollable container
//   - outside scrollbar
//   - scrollable horizontal
//   - stylable bar
//     - round corners
//     - border/bg
//   - scrollable by moving the scrollbar
//   - (jump-)scrollable by clicking the scrollbararea
//   - minimum size for scrollbar
// - canvas
//   - remove dependency from driver to actual implementations (fyne.Container or even worse widget.ScrollContainer)
//     -> the driver should probably use a fyne-interface

type mainGrid struct {
	baseWidget
	content         fyne.CanvasObject
	contentSelector fyne.CanvasObject
	controls        fyne.CanvasObject
	status          fyne.CanvasObject
	statusBar       fyne.CanvasObject
}

func NewMainGrid(
	content fyne.CanvasObject,
	controls fyne.CanvasObject,
	status fyne.CanvasObject,
	statusBar fyne.CanvasObject,
) fyne.Widget {
	return &mainGrid{
		content:   content,
		controls:  controls,
		status:    status,
		statusBar: statusBar,
	}
}

func (g *mainGrid) CreateRenderer() fyne.WidgetRenderer {
	return &mainGridRenderer{
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{
			g.content,
			g.controls,
			g.status,
			g.statusBar,
		}},
		grid: g,
	}
}

func (g *mainGrid) Hide() {
	g.hide(g)
}

func (g *mainGrid) MinSize() fyne.Size {
	return g.minSize(g)
}

func (g *mainGrid) Resize(size fyne.Size) {
	g.resize(g, size)
}

func (g *mainGrid) Show() {
	g.show(g)
}

type mainGridRenderer struct {
	baseRenderer
	grid *mainGrid
}

func (r *mainGridRenderer) Layout(size fyne.Size) {
	stHeight := r.grid.status.MinSize().Height
	ctrlSize := r.grid.controls.MinSize()
	r.grid.controls.Move(fyne.NewPos(0, (stHeight-ctrlSize.Height)/2))
	r.grid.controls.Resize(ctrlSize)

	r.grid.status.Move(fyne.NewPos(ctrlSize.Width, 0))
	r.grid.status.Resize(fyne.NewSize(size.Width-ctrlSize.Width, stHeight))

	sbHeight := r.grid.statusBar.MinSize().Height
	r.grid.statusBar.Move(fyne.NewPos(0, size.Height-sbHeight))
	r.grid.statusBar.Resize(fyne.NewSize(size.Width, sbHeight))

	headerHeight := fyne.Max(ctrlSize.Height, stHeight)
	r.grid.content.Move(fyne.NewPos(0, headerHeight+theme.Padding()))
	r.grid.content.Resize(fyne.NewSize(size.Width, size.Height-headerHeight-sbHeight-theme.Padding()))
}

func (r *mainGridRenderer) MinSize() fyne.Size {
	ctrlMinSize := r.grid.controls.MinSize()
	sMinSize := r.grid.status.MinSize()
	cMinSize := r.grid.content.MinSize()
	sbMinSize := r.grid.statusBar.MinSize()
	return fyne.NewSize(
		fyne.Max(cMinSize.Width, ctrlMinSize.Width+sMinSize.Width),
		sbMinSize.Height+cMinSize.Height+fyne.Max(ctrlMinSize.Height, sMinSize.Height)+theme.Padding(),
	)
}

func (r *mainGridRenderer) Refresh() {
}
