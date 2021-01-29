package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// TODO: checken, ob erben von widget.Button sinnvoll ist
// TODO
// - Tooltip on hover in Mauszeigernähe (Hover-Event-Pos -> Hover-Event braucht delay -> ggf. spezielles Hover-Event)
// - Rahmen
// - Rahmen on Hover
// - kein Hover-Rahmen bei disabled
// - Klick-Effekt (z.B. kurze Zeit (100ms) disabled style)
// - Hover-Effekt (hilight) -> ggf. disjunkt zum hover-Rahmen
// - runde Kanten

type iconButton struct {
	widget.BaseWidget
	badgeCount int
	disabled   bool
	hovered    bool
	icon       fyne.Resource
	iconSize   fyne.Size
	onTap      func()
	pad        bool
}

func newIconButton(icon fyne.Resource, onTap func()) *iconButton {
	button := &iconButton{icon: icon, onTap: onTap, iconSize: fyne.NewSize(48, 48), pad: true}
	button.ExtendBaseWidget(button)
	return button
}

func (b *iconButton) CreateRenderer() fyne.WidgetRenderer {
	return newIconButtonRenderer(b)
}

func (b *iconButton) Disabled() bool {
	return b.disabled
}

func (b *iconButton) Disable() {
	if !b.disabled {
		b.disabled = true
		b.Refresh()
	}
}

func (b *iconButton) Enable() {
	if b.disabled {
		b.disabled = false
		b.Refresh()
	}
}

func (b *iconButton) MouseIn(_ *desktop.MouseEvent) {
	b.hovered = true
	b.Refresh()
}

func (b *iconButton) MouseOut() {
	b.hovered = false
	b.Refresh()
}

func (b *iconButton) MouseMoved(_ *desktop.MouseEvent) {
}

func (b *iconButton) Tapped(*fyne.PointEvent) {
	if b.disabled {
		return
	}
	b.onTap()
}

func (b *iconButton) TappedSecondary(*fyne.PointEvent) {
}

func (b *iconButton) UpdateBadgeCount(count int) {
	b.badgeCount = count
	b.Refresh()
}

type iconButtonRenderer struct {
	baseRenderer
	badgeBackgroundLeft   *canvas.Circle
	badgeBackgroundMiddle *canvas.Rectangle
	badgeBackgroundRight  *canvas.Circle
	badgeText             *canvas.Text
	background            *canvas.Rectangle
	button                *iconButton
	disabledIcon          *canvas.Image
	icon                  *canvas.Image
}

func newIconButtonRenderer(b *iconButton) *iconButtonRenderer {
	var icon *canvas.Image
	icon = canvas.NewImageFromResource(b.icon)
	badgeBGColor := &color.RGBA{R: 200, A: 255}
	badgeBGL := canvas.NewCircle(badgeBGColor)
	badgeBGL.Hide()
	badgeBGM := canvas.NewRectangle(badgeBGColor)
	badgeBGM.Hide()
	badgeBGR := canvas.NewCircle(badgeBGColor)
	badgeBGR.Hide()
	badgeText := canvas.NewText("0", color.White)
	badgeText.TextStyle.Bold = true
	badgeText.Hide()
	bg := &canvas.Rectangle{FillColor: color.Transparent}
	objects := []fyne.CanvasObject{bg, icon, badgeBGL, badgeBGM, badgeBGR, badgeText}
	return &iconButtonRenderer{
		badgeBackgroundLeft:   badgeBGL,
		badgeBackgroundMiddle: badgeBGM,
		badgeBackgroundRight:  badgeBGR,
		badgeText:             badgeText,
		background:            bg,
		baseRenderer:          baseRenderer{objects: objects},
		button:                b,
		disabledIcon:          canvas.NewImageFromResource(theme.NewDisabledResource(b.icon)),
		icon:                  icon,
	}

}

func (r *iconButtonRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)

	if r.button.pad {
		size = size.Subtract(fyne.NewSize(theme.Padding()*2, theme.Padding()*2))
		r.icon.Move(fyne.NewPos(theme.Padding(), theme.Padding()))
		r.disabledIcon.Move(fyne.NewPos(theme.Padding(), theme.Padding()))
	}
	r.icon.Resize(size)
	r.disabledIcon.Resize(size)

	r.badgeText.TextSize = r.icon.Size().Height * 0.4
	r.badgeText.Resize(r.badgeText.MinSize())
	badgeHeight := r.badgeText.MinSize().Height
	badgeWidth := r.badgeText.MinSize().Width + 6
	var textOffset float32 = 3
	if badgeWidth < badgeHeight {
		textOffset += badgeHeight - badgeWidth
		badgeWidth = badgeHeight
	}
	circleSize := fyne.NewSize(badgeHeight, badgeHeight)
	rectWidth := badgeWidth - badgeHeight
	rectSize := fyne.NewSize(rectWidth, badgeHeight)
	r.badgeBackgroundLeft.Resize(circleSize)
	r.badgeBackgroundMiddle.Resize(rectSize)
	r.badgeBackgroundRight.Resize(circleSize)
	badgePos := fyne.NewPos(
		r.icon.Position().X+r.icon.Size().Width-badgeWidth,
		r.icon.Position().Y,
	)
	badgeTextPos := fyne.NewPos(
		badgePos.X+textOffset,
		badgePos.Y,
	)
	r.badgeText.Move(badgeTextPos)
	r.badgeBackgroundLeft.Move(badgePos)
	r.badgeBackgroundMiddle.Move(badgePos.Add(fyne.NewPos(badgeHeight/2, 0)))
	r.badgeBackgroundRight.Move(badgePos.Add(fyne.NewPos(rectWidth, 0)))
}

func (r *iconButtonRenderer) MinSize() fyne.Size {
	if r.button.pad {
		return fyne.NewSize(theme.Padding()*2, theme.Padding()*2).Add(r.button.iconSize)
	}
	return r.button.iconSize
}

func (r *iconButtonRenderer) Refresh() {
	if r.button.hovered {
		r.background.FillColor = theme.HoverColor()
	} else {
		r.background.FillColor = color.Transparent
	}
	if r.button.Disabled() {
		r.objects[1] = r.disabledIcon
	} else {
		r.objects[1] = r.icon
	}
	if r.button.badgeCount > 0 {
		if r.button.badgeCount > 9500 {
			r.badgeText.Text = ">9K"
		} else if r.button.badgeCount > 999 {
			r.badgeText.Text = fmt.Sprintf("%dK", r.button.badgeCount/1000)
		} else {
			r.badgeText.Text = fmt.Sprintf("%d", r.button.badgeCount)
		}
		r.badgeBackgroundLeft.Show()
		r.badgeBackgroundMiddle.Show()
		r.badgeBackgroundRight.Show()
		r.badgeText.Show()
	} else {
		r.badgeBackgroundLeft.Hide()
		r.badgeBackgroundMiddle.Hide()
		r.badgeBackgroundRight.Hide()
		r.badgeText.Hide()
	}
	canvas.Refresh(r.button)
}
