module github.com/toaster/tmpc

go 1.12

replace fyne.io/fyne => ../fyne

replace github.com/fhs/gompd => ../gompd

replace github.com/romantomjak/shoutcast => ../shoutcast

require (
	fyne.io/fyne v1.0.1
	github.com/fhs/gompd v2.0.0+incompatible
	github.com/fhs/gompd/v2 v2.0.3 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20190328170749-bb2674552d8f // indirect
	github.com/hajimehoshi/go-mp3 v0.2.0
	github.com/hajimehoshi/oto v0.3.2
	github.com/pkg/errors v0.8.1
	github.com/romantomjak/shoutcast v1.1.0
	golang.org/x/net v0.0.0-20190724013045-ca1201d0de80
)
