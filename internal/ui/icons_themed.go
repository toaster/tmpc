package ui

import "fyne.io/fyne/theme"

var AlbumIcon *theme.ThemedResource

func init() {
	AlbumIcon = theme.NewThemedResource(rscAlbum, nil)
}

var ErrorIcon *theme.ThemedResource

func init() {
	ErrorIcon = theme.NewThemedResource(rscError, nil)
}

var HeadphonesIcon *theme.ThemedResource

func init() {
	HeadphonesIcon = theme.NewThemedResource(rscHeadphones, nil)
}

var ListIcon *theme.ThemedResource

func init() {
	ListIcon = theme.NewThemedResource(rscList, nil)
}

var MusicIcon *theme.ThemedResource

func init() {
	MusicIcon = theme.NewThemedResource(rscMusic, nil)
}

var NextIcon *theme.ThemedResource

func init() {
	NextIcon = theme.NewThemedResource(rscNext, nil)
}

var NoMusicIcon *theme.ThemedResource

func init() {
	NoMusicIcon = theme.NewThemedResource(rscNoMusic, nil)
}

var PauseIcon *theme.ThemedResource

func init() {
	PauseIcon = theme.NewThemedResource(rscPause, nil)
}

var PlayShadowIcon *theme.ThemedResource

func init() {
	PlayShadowIcon = theme.NewThemedResource(rscPlayShadow, nil)
}

var PlayIcon *theme.ThemedResource

func init() {
	PlayIcon = theme.NewThemedResource(rscPlay, nil)
}

var PluggedIcon *theme.ThemedResource

func init() {
	PluggedIcon = theme.NewThemedResource(rscPlugged, nil)
}

var PrevIcon *theme.ThemedResource

func init() {
	PrevIcon = theme.NewThemedResource(rscPrev, nil)
}

var QueueIcon *theme.ThemedResource

func init() {
	QueueIcon = theme.NewThemedResource(rscQueue, nil)
}

var StopIcon *theme.ThemedResource

func init() {
	StopIcon = theme.NewThemedResource(rscStop, nil)
}

var UnpluggedIcon *theme.ThemedResource

func init() {
	UnpluggedIcon = theme.NewThemedResource(rscUnplugged, nil)
}
