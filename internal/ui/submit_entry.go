package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// SubmitEntry is an entry that can be submitted by pressing Enter.
type SubmitEntry struct {
	widget.Entry
	onSubmit func(string)
}

// NewSubmitEntry creates a SubmitEntry.
func NewSubmitEntry(onSubmit func(string), icon fyne.Resource) *SubmitEntry {
	e := &SubmitEntry{onSubmit: onSubmit}
	e.ExtendBaseWidget(e)
	if icon != nil {
		submitButton := widget.NewButton("", func() { e.onSubmit(e.Text) })
		submitButton.Importance = widget.LowImportance
		submitButton.SetIcon(icon)
		e.ActionItem = submitButton
	}
	return e
}

// TypedKey handles the submit.
func (e *SubmitEntry) TypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyReturn, fyne.KeyEnter:
		e.onSubmit(e.Text)
	default:
		e.Entry.TypedKey(key)
	}
}
