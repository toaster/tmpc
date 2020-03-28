package ui

import (
	"fyne.io/fyne"
	"fyne.io/fyne/widget"
)

// SubmitEntry is an entry that can be submitted by pressing Enter.
type SubmitEntry struct {
	widget.Entry
	onSubmit func(string)
}

// NewSubmitEntry creates a SubmitEntry.
func NewSubmitEntry(onSubmit func(string)) *SubmitEntry {
	e := &SubmitEntry{onSubmit: onSubmit}
	e.ExtendBaseWidget(e)
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
