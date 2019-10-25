package ui

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

type progressBar struct {
	baseWidget
	cur     int
	max     int
	min     int
	onClick func(int)
}

func NewProgressBar(min, max int, onClick func(int)) *progressBar {
	return &progressBar{
		cur:     min,
		max:     max,
		min:     min,
		onClick: onClick,
	}
}

func (p *progressBar) CreateRenderer() fyne.WidgetRenderer {
	background := canvas.NewRectangle(theme.BackgroundColor())
	border := canvas.NewRectangle(theme.ButtonColor())
	bar := canvas.NewRectangle(theme.PrimaryColor())
	return &progressBarRenderer{
		baseRenderer: baseRenderer{objects: []fyne.CanvasObject{border, background, bar}},
		bar:          bar,
		bg:           background,
		border:       border,
		p:            p,
	}
}

func (p *progressBar) MinSize() fyne.Size {
	return p.minSize(p)
}

func (p *progressBar) ReInit(min, max, cur int) {
	p.min = min
	p.max = max
	p.cur = cur
	widget.Refresh(p)
}

func (p *progressBar) Resize(size fyne.Size) {
	p.resize(p, size)
}

func (p *progressBar) Tapped(e *fyne.PointEvent) {
	if p.onClick != nil && p.max != p.min {
		offset := fyne.NewPos(theme.Padding()+1, theme.Padding()+1)
		size := p.Size().Subtract(fyne.NewSize(theme.Padding()*2-2, theme.Padding()*2-2))
		if eventIsInArea(e, offset, size) {
			pos := e.Position.Subtract(offset)
			ratio := float64(pos.X) / float64(size.Width-2)
			cur := int(ratio*float64(p.max-p.min)) + p.min
			if cur > p.max-p.min {
				cur = p.max - p.min
			} else if cur < 0 {
				cur = 0
			}
			p.onClick(cur)
		}
	}
}

func (p *progressBar) TappedSecondary(*fyne.PointEvent) {
}

func (p *progressBar) Update(cur int) {
	if cur > p.max {
		cur = p.max
	} else if cur < p.min {
		cur = p.min
	}
	p.cur = cur
	widget.Refresh(p)
}

type progressBarRenderer struct {
	baseRenderer
	border *canvas.Rectangle
	bg     *canvas.Rectangle
	bar    *canvas.Rectangle
	p      *progressBar
}

func (p *progressBarRenderer) ApplyTheme() {
	p.bar.FillColor = theme.PrimaryColor()

	p.Refresh()
}

func (r *progressBarRenderer) Layout(size fyne.Size) {
	pos := fyne.NewPos(theme.Padding(), theme.Padding())
	r.border.Move(pos)
	r.bg.Move(pos.Add(fyne.NewPos(1, 1)))
	r.bar.Move(pos.Add(fyne.NewPos(2, 2)))

	width := size.Width - theme.Padding()*2
	r.border.Resize(fyne.NewSize(width, 10))
	r.bg.Resize(fyne.NewSize(width-2, 8))

	var ratio float64
	if r.p.max != r.p.min {
		ratio = float64(r.p.cur-r.p.min) / float64(r.p.max-r.p.min)
	}
	r.bar.Resize(fyne.NewSize(int(ratio*float64(width-4)), 6))
}

func (r *progressBarRenderer) MinSize() fyne.Size {
	return fyne.NewSize(10+theme.Padding()*2, 10+theme.Padding()*2)
}

func (r *progressBarRenderer) Refresh() {
	r.Layout(r.p.Size())
}
