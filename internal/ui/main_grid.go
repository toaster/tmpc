package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type mainGrid struct {
	widget.BaseWidget
	content   fyne.CanvasObject
	controls  fyne.CanvasObject
	status    fyne.CanvasObject
	statusBar fyne.CanvasObject
}

// NewMainGrid returns a new container for the content of the TMPC main window.
func NewMainGrid(
	content fyne.CanvasObject,
	controls fyne.CanvasObject,
	status fyne.CanvasObject,
	statusBar fyne.CanvasObject,
) fyne.Widget {
	grid := &mainGrid{
		content:   content,
		controls:  controls,
		status:    status,
		statusBar: statusBar,
	}
	grid.ExtendBaseWidget(grid)
	return grid
}

func (g *mainGrid) CreateRenderer() fyne.WidgetRenderer {
	separator := canvas.NewRectangle(theme.PlaceHolderColor())
	return &mainGridRenderer{
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{
			g.content,
			g.controls,
			separator,
			g.status,
			g.statusBar,
		}},
		grid:      g,
		separator: separator,
	}
}

type mainGridRenderer struct {
	baseRenderer
	grid      *mainGrid
	separator fyne.CanvasObject
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

	headerContentHeight := fyne.Max(ctrlSize.Height, stHeight) + theme.Padding()
	r.separator.Resize(fyne.NewSize(size.Width, 1))
	r.separator.Move(fyne.NewPos(0, headerContentHeight))
	headerHeight := headerContentHeight + r.separator.Size().Height
	r.grid.content.Move(fyne.NewPos(0, headerHeight))
	r.grid.content.Resize(fyne.NewSize(size.Width, size.Height-headerHeight-sbHeight))
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
	canvas.Refresh(r.grid)
}
