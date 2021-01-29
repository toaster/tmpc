module github.com/toaster/tmpc

go 1.12

replace fyne.io/fyne => ../fyne

replace github.com/fhs/gompd => ../gompd

replace github.com/romantomjak/shoutcast => ../shoutcast

require (
	fyne.io/fyne v1.0.1
	github.com/fhs/gompd v2.0.0+incompatible
	github.com/fhs/gompd/v2 v2.0.3 // indirect
	github.com/hajimehoshi/go-mp3 v0.3.0
	github.com/hajimehoshi/oto v0.6.1
	github.com/pkg/errors v0.9.1
	github.com/romantomjak/shoutcast v1.1.0
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
)
