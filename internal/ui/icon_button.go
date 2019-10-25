package ui

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

// TODO
// - Tooltip on hover in Mauszeigernähe (Hover-Event-Pos -> Hover-Event braucht delay -> ggf. spezielles Hover-Event)
// - Rahmen
// - Rahmen on Hover
// - kein Hover-Rahmen bei disabled
// - Klick-Effekt (z.B. kurze Zeit (100ms) disabled style)
// - Hover-Effekt (hilight) -> ggf. disjunkt zum hover-Rahmen
// - runde Kanten

type iconButton struct {
	baseWidget
	disabled bool
	hovered  bool
	icon     fyne.Resource
	iconSize fyne.Size
	onTap    func()
	pad      bool
}

func NewIconButton(icon fyne.Resource, onTap func()) *iconButton {
	return &iconButton{icon: icon, onTap: onTap, iconSize: fyne.NewSize(48, 48), pad: true}
}

func (b *iconButton) CreateRenderer() fyne.WidgetRenderer {
	return newIconButtonRenderer(b)
}

func (b *iconButton) Disabled() bool {
	return b.disabled
}

func (b *iconButton) Hide() {
	b.hide(b)
}

func (b *iconButton) MinSize() fyne.Size {
	return b.minSize(b)
}

func (b *iconButton) Resize(size fyne.Size) {
	b.resize(b, size)
}

func (b *iconButton) Show() {
	b.show(b)
}

func (b *iconButton) Disable() {
	if !b.disabled {
		b.disabled = true
		widget.Refresh(b)
	}
}

func (b *iconButton) Enable() {
	if b.disabled {
		b.disabled = false
		widget.Refresh(b)
	}
}

func (b *iconButton) MouseIn(e *desktop.MouseEvent) {
	b.hovered = true
	widget.Refresh(b)
}

func (b *iconButton) MouseOut() {
	b.hovered = false
	widget.Refresh(b)
}

func (b *iconButton) MouseMoved(e *desktop.MouseEvent) {
}

func (b *iconButton) Tapped(*fyne.PointEvent) {
	if b.disabled {
		return
	}
	b.onTap()
}

func (b *iconButton) TappedSecondary(*fyne.PointEvent) {
}

type iconButtonRenderer struct {
	baseRenderer
	button       *iconButton
	disabledIcon *canvas.Image
	icon         *canvas.Image
}

func newIconButtonRenderer(b *iconButton) *iconButtonRenderer {
	var icon *canvas.Image
	objects := []fyne.CanvasObject{}
	icon = canvas.NewImageFromResource(b.icon)
	objects = append(objects, icon)
	return &iconButtonRenderer{
		baseRenderer: baseRenderer{objects: objects},
		button:       b,
		disabledIcon: canvas.NewImageFromResource(theme.NewDisabledResource(b.icon)),
		icon:         icon,
	}

}

func (b *iconButtonRenderer) BackgroundColor() color.Color {
	if b.button.hovered {
		return theme.HoverColor()
	} else {
		return b.baseRenderer.BackgroundColor()
	}
}

func (r *iconButtonRenderer) Layout(size fyne.Size) {
	if r.button.pad {
		size = size.Subtract(fyne.NewSize(theme.Padding()*2, theme.Padding()*2))
		r.icon.Move(fyne.NewPos(theme.Padding(), theme.Padding()))
		r.disabledIcon.Move(fyne.NewPos(theme.Padding(), theme.Padding()))
	}
	r.icon.Resize(size)
	r.disabledIcon.Resize(size)
}

func (r *iconButtonRenderer) MinSize() fyne.Size {
	if r.button.pad {
		return fyne.NewSize(theme.Padding()*2, theme.Padding()*2).Add(r.button.iconSize)
	}
	return r.button.iconSize
}

func (r *iconButtonRenderer) Refresh() {
	if r.button.Disabled() {
		r.objects = []fyne.CanvasObject{r.disabledIcon}
	} else {
		r.objects = []fyne.CanvasObject{r.icon}
	}
	canvas.Refresh(r.button)
}
