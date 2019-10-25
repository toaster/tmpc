package ui

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/theme"
)

type aTheme struct {
	base fyne.Theme
}

func NewTheme() fyne.Theme {
	return aTheme{theme.DarkTheme()}
}

func (t aTheme) BackgroundColor() color.Color {
	return t.base.BackgroundColor()
}

func (t aTheme) ButtonColor() color.Color {
	return t.base.ButtonColor()
}

func (t aTheme) DisabledButtonColor() color.Color {
	return t.base.DisabledButtonColor()
}

func (t aTheme) DisabledIconColor() color.Color {
	return t.base.DisabledIconColor()
}

func (t aTheme) DisabledTextColor() color.Color {
	return t.base.DisabledTextColor()
}

func (t aTheme) FocusColor() color.Color {
	return t.base.FocusColor()
}

func (t aTheme) HoverColor() color.Color {
	return t.base.HoverColor()
}

func (t aTheme) HyperlinkColor() color.Color {
	return t.base.HyperlinkColor()
}

func (t aTheme) IconColor() color.Color {
	return t.base.IconColor()
}

func (t aTheme) IconInlineSize() int {
	return t.base.IconInlineSize()
}

func (t aTheme) Padding() int {
	return 1
}

func (t aTheme) PlaceHolderColor() color.Color {
	return t.base.PlaceHolderColor()
}

func (t aTheme) PrimaryColor() color.Color {
	return t.base.PrimaryColor()
}

func (t aTheme) ScrollBarColor() color.Color {
	return t.base.ScrollBarColor()
}

func (t aTheme) ScrollBarSize() int {
	return 10
}

func (t aTheme) ScrollBarSmallSize() int {
	return 3
}

func (t aTheme) ShadowColor() color.Color {
	return t.base.ShadowColor()
}

func (t aTheme) TextBoldFont() fyne.Resource {
	return t.base.TextBoldFont()
}

func (t aTheme) TextBoldItalicFont() fyne.Resource {
	return t.base.TextBoldItalicFont()
}

func (t aTheme) TextColor() color.Color {
	return t.base.TextColor()
}

func (t aTheme) TextFont() fyne.Resource {
	return t.base.TextFont()
}

func (t aTheme) TextItalicFont() fyne.Resource {
	return t.base.TextItalicFont()
}

func (t aTheme) TextMonospaceFont() fyne.Resource {
	return t.base.TextMonospaceFont()
}

func (t aTheme) TextSize() int {
	return 12
}
