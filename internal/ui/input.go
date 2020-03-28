package ui

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func textMinSize(text string, size int, style fyne.TextStyle) fyne.Size {
	t := canvas.NewText(text, color.Black)
	t.TextSize = size
	t.TextStyle = style
	return t.MinSize()
}

type inputRenderer struct {
	icon  *widget.Icon
	label *canvas.Text

	objects []fyne.CanvasObject
	combo   *Input
}

func (r *inputRenderer) MinSize() fyne.Size {
	min := textMinSize(r.combo.PlaceHolder, r.label.TextSize, r.label.TextStyle)

	for _, option := range r.combo.Options {
		optionMin := textMinSize(option, r.label.TextSize, r.label.TextStyle)
		min = min.Union(optionMin)
	}

	min = min.Add(fyne.NewSize(theme.Padding()*4, theme.Padding()*2))
	return min.Add(fyne.NewSize(theme.IconInlineSize()+theme.Padding(), 0))
}

// Layout the components of the button widget
func (r *inputRenderer) Layout(size fyne.Size) {
	inner := size.Subtract(fyne.NewSize(theme.Padding()*4, theme.Padding()*2))

	offset := fyne.NewSize(theme.IconInlineSize(), 0)
	labelSize := inner.Subtract(offset)
	r.label.Resize(labelSize)
	r.label.Move(fyne.NewPos(theme.Padding()*2, theme.Padding()))

	r.icon.Resize(fyne.NewSize(theme.IconInlineSize(), theme.IconInlineSize()))
	r.icon.Move(fyne.NewPos(
		size.Width-theme.IconInlineSize()-theme.Padding()*2,
		(size.Height-theme.IconInlineSize())/2))
}

func (r *inputRenderer) BackgroundColor() color.Color {
	if r.combo.hovered {
		return theme.HoverColor()
	}
	return theme.ButtonColor()
}

func (r *inputRenderer) Refresh() {
	r.label.Color = theme.TextColor()
	r.label.TextSize = theme.TextSize()

	if r.combo.Selected == "" {
		r.label.Text = r.combo.PlaceHolder
	} else {
		r.label.Text = r.combo.Selected
	}

	if false { // r.combo.down {
		r.icon.Resource = theme.MenuDropUpIcon()
	} else {
		r.icon.Resource = theme.MenuDropDownIcon()
	}

	r.Layout(r.combo.Size())
	canvas.Refresh(r.combo)
}

func (r *inputRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *inputRenderer) Destroy() {
	if r.combo.popUp != nil {
		c := fyne.CurrentApp().Driver().CanvasForObject(r.combo)
		c.SetOverlay(nil)
		// TODO: WTF?
		// cache.Renderer(r.combo.popUp).Destroy()
		// thus?: widget.Renderer(r.combo.popUp).Destroy()
		r.combo.popUp = nil
	}
}

// Input widget has a list of options, with the current one shown, and triggers an event func when clicked
type Input struct {
	widget.BaseWidget

	Selected    string
	Options     []string
	PlaceHolder string
	OnChanged   func(string) `json:"-"`

	hovered bool
	popUp   *widget.PopUp
}

// Resize sets a new size for a widget.
// Note this should not be used if the widget is being managed by a Layout within a Container.
func (i *Input) Resize(size fyne.Size) {
	i.BaseWidget.Resize(size)

	if i.popUp != nil {
		i.popUp.Content.Resize(fyne.NewSize(size.Width, i.popUp.MinSize().Height))
	}
}

func (i *Input) optionTapped(text string) {
	i.SetSelected(text)
	i.popUp = nil
}

// Tapped is called when a pointer tapped event is captured and triggers any tap handler
func (i *Input) Tapped(*fyne.PointEvent) {
	// TODO: CanvasForObject only works for final widget (no one can extend this widget then)
	c := fyne.CurrentApp().Driver().CanvasForObject(i)

	var items []*fyne.MenuItem
	for _, option := range i.Options {
		text := option // capture
		item := fyne.NewMenuItem(option, func() {
			i.optionTapped(text)
		})
		items = append(items, item)
	}

	// TODO: AbsolutePositionForObject only works for final widget (no one can extend this widget then)
	buttonPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(i)
	popUpPos := buttonPos.Add(fyne.NewPos(0, i.Size().Height))

	i.popUp = widget.NewPopUpMenuAtPosition(fyne.NewMenu("", items...), c, popUpPos)
	i.popUp.Resize(fyne.NewSize(i.Size().Width, i.popUp.Content.MinSize().Height))
}

// TappedSecondary is called when a secondary pointer tapped event is captured
func (i *Input) TappedSecondary(*fyne.PointEvent) {
}

// MouseIn is called when a desktop pointer enters the widget
func (i *Input) MouseIn(*desktop.MouseEvent) {
	i.hovered = true
	i.Refresh()
}

// MouseOut is called when a desktop pointer exits the widget
func (i *Input) MouseOut() {
	i.hovered = false
	i.Refresh()
}

// MouseMoved is called when a desktop pointer hovers over the widget
func (i *Input) MouseMoved(*desktop.MouseEvent) {
}

// MinSize returns the size that this widget should not shrink below
func (i *Input) MinSize() fyne.Size {
	i.ExtendBaseWidget(i)
	return i.BaseWidget.MinSize()
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (i *Input) CreateRenderer() fyne.WidgetRenderer {
	i.ExtendBaseWidget(i)
	icon := widget.NewIcon(theme.MenuDropDownIcon())
	text := canvas.NewText(i.Selected, theme.TextColor())

	if i.Selected == "" {
		text.Text = i.PlaceHolder
	}
	text.Alignment = fyne.TextAlignLeading

	objects := []fyne.CanvasObject{
		text, icon,
	}

	return &inputRenderer{icon, text, objects, i}
}

// SetSelected sets the current option of the select widget
func (i *Input) SetSelected(text string) {
	for _, option := range i.Options {
		if text == option {
			i.Selected = text
		}
	}

	if i.OnChanged != nil {
		i.OnChanged(text)
	}

	i.Refresh()
}

// NewSelect creates a new select widget with the set list of options and changes handler
func NewSelect(options []string, changed func(string)) *Input {
	return &Input{widget.BaseWidget{}, "", options, "", changed, false, nil}
}
