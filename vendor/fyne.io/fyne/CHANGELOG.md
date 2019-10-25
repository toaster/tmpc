# Changelog

This file lists the main changes with each version of the Fyne toolkit.
More detailed release notes can be found on the [releases page](https://github.com/fyne-io/fyne/releases). 

## 1.1 - (in progress)

### Added

* Menubar and PopUpMenu (#41)
* PopUp widgets (regular and modal) and canvas overlay support (#242)
* Add shadow support for overlays, buttons and scrollcontainer
* Text can now be selected (#67)
* Support moving through inputs with Tab / Shift-Tab (#82)
* canvas.Capture() to save the content of a canvas
* Horizontal layout for widget.Radio
* Select widget (#21)
* Add support for disabling widgets (#234)
* Support for changing icon color (#246)
* Button hover effect
* Pointer drag event to main API
* support for desktop mouse move events
* Add a new "hints" build tag that can suggest UI improvements

### Changed

* TabContainer tab location can now be set with SetTabLocation()
* Dialog windows now appear as modal popups within a window
* Don't add a button bar to a form if it has no buttons
* Moved driver/gl package to internal/driver/gl
* Clicking/Tapping in an entry will position the cursor
* A container with no layout will not change the position or size of it's content

### Fixed

* Allow scrollbars to be dragged (#133)
* Unicode char input with Option key on macOS (#247)
* Resizng fixed size windows (#248)
* Fixed various bugs in window sizing and padding
* Button icons do not center align if label is empty (#284)


## 1.0.1 - 20 April 2019

### Added

* Support for go modules
* Transparent backgrounds for widgets
* Entry.OnCursorChanged()
* Radio.Append() and Radio.SetSelected() (#229)

### Changed

* Clicking outside a focused element will unfocus it
* Handle key repeat for non-runes (#165)

### Fixed

* Remove duplicate options from a Radio widget (#230)
* Issue where paste shortcut is not called for Ctrl-V keyboard combination
* Cursor position when clearing text in Entry (#214)
* Antialias of lines and circles (fyne-io/examples#14)
* Crash on centering of windows (#220)
* Possible crash when closing secondary windows
* Possible crash when showing dialog
* Initial visibility of scroll bar in ScrollContainer
* Setting window icon when different from app icon.
* Possible panic on app.Quit() (#175)
* Various caches and race condition issues (#194, #217, #209).


## 1.0 19 - March 2019

The first major release of the Fyne toolkit delivers a stable release of the
main functionality required to build basic GUI applications across multiple
platforms.

### Features

* Canvas API (rect, line, circle, text, image)
* Widget API (box, button, check, entry, form, group, hyperlink, icon, label, progress bar, radio, scroller, tabs and toolbar)
* Light and dark themes
* Pointer, key and shortcut APIs (generic and desktop extension)
* OpenGL driver for Linux, macOS and Windows
* Tools for embedding data and packaging releases

