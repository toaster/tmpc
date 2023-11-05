package ui

import (
	"embed"
	"io"

	"fyne.io/fyne/v2"
)

//go:embed resources
var resources embed.FS

var rscPauseIndicator = loadResource("pause_indicator.svg")
var rscPlayIndicator = loadResource("play_indicator.svg")
var rscStopIndicator = loadResource("stop_indicator.svg")

func loadResource(name string) fyne.Resource {
	f, err := resources.Open("resources/" + name)
	if err != nil {
		fyne.LogError("failed to load resource", err)
		return nil
	}

	defer func() {
		_ = f.Close()
	}()

	bytes, err := io.ReadAll(f)
	if err != nil {
		fyne.LogError("failed to load resource", err)
		return nil
	}

	return fyne.NewStaticResource(name, bytes)
}
