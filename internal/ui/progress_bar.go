package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type progressBar struct {
	widget.BaseWidget
	cur     int
	max     int
	min     int
	onClick func(int)
}

func newProgressBar(min, max int, onClick func(int)) *progressBar {
	b := &progressBar{
		cur:     min,
		max:     max,
		min:     min,
		onClick: onClick,
	}
	b.ExtendBaseWidget(b)
	return b
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

func (p *progressBar) ReInit(min, max, cur int) {
	p.min = min
	p.max = max
	p.cur = cur
	p.Refresh()
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
	p.Refresh()
}

type progressBarRenderer struct {
	baseRenderer
	border *canvas.Rectangle
	bg     *canvas.Rectangle
	bar    *canvas.Rectangle
	p      *progressBar
}

func (r *progressBarRenderer) Layout(size fyne.Size) {
	const inset = 1
	const barHeight = 6
	const bgHeight = barHeight + 2*inset
	const borderHeight = bgHeight + 2*inset

	pos := fyne.NewPos(theme.Padding(), theme.Padding())
	r.border.Move(pos)
	r.bg.Move(pos.Add(fyne.NewPos(inset, inset)))
	r.bar.Move(pos.Add(fyne.NewPos(2*inset, 2*inset)))

	width := size.Width - theme.Padding()*2
	r.border.Resize(fyne.NewSize(width, borderHeight))
	r.bg.Resize(fyne.NewSize(width-2*inset, bgHeight))

	var ratio float64
	if r.p.max != r.p.min {
		ratio = float64(r.p.cur-r.p.min) / float64(r.p.max-r.p.min)
	}
	r.bar.Resize(fyne.NewSize(float32(ratio)*(width-4*inset), barHeight))
}

func (r *progressBarRenderer) MinSize() fyne.Size {
	return fyne.NewSize(10+theme.Padding()*2, 10+theme.Padding()*2)
}

func (r *progressBarRenderer) Refresh() {
	r.bar.FillColor = theme.PrimaryColor()
	r.bg.FillColor = theme.BackgroundColor()
	r.border.FillColor = theme.HoverColor()
	r.bar.Refresh()
	r.bg.Refresh()
	r.border.Refresh()
	r.Layout(r.p.Size())
}
