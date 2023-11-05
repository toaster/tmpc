package ui

import (
	"embed"
	"io"
	"io/fs"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

//go:embed icons
var icons embed.FS

//go:embed resources
var resources embed.FS

var (
	IconAlbum = loadIcon("album.svg")
	IconList  = loadIcon("list.svg")
	IconQueue = loadIcon("queue.svg")

	iconError      = loadIcon("error.svg")
	iconHeadphones = loadIcon("headphones.svg")
	iconMusic      = loadIcon("music.svg")
	iconNext       = loadIcon("next.svg")
	iconNoMusic    = loadIcon("no_music.svg")
	iconPause      = loadIcon("pause.svg")
	iconPlay       = loadIcon("play.svg")
	// iconPlayShadow = loadIcon("play-shadow.svg")
	iconPlugged   = loadIcon("plugged.svg")
	iconPrev      = loadIcon("prev.svg")
	iconStop      = loadIcon("stop.svg")
	iconUnplugged = loadIcon("unplugged.svg")
)

var (
	rscPauseIndicator = loadResource("pause_indicator.svg")
	rscPlayIndicator  = loadResource("play_indicator.svg")
	rscStopIndicator  = loadResource("stop_indicator.svg")
)

func loadIcon(name string) fyne.Resource {
	return theme.NewThemedResource(loadResourceFromFS(name, icons, "icon"))
}

func loadResource(name string) fyne.Resource {
	return loadResourceFromFS(name, resources, "resource")
}

func loadResourceFromFS(name string, fs fs.FS, typ string) fyne.Resource {
	f, err := fs.Open(typ + "s/" + name)
	if err != nil {
		fyne.LogError("failed to load "+typ, err)
		return nil
	}

	defer func() {
		_ = f.Close()
	}()

	bytes, err := io.ReadAll(f)
	if err != nil {
		fyne.LogError("failed to load "+typ, err)
		return nil
	}

	return fyne.NewStaticResource(name, bytes)
}
