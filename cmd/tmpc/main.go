package main

import (
	"flag"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"runtime/pprof"

	"fyne.io/fyne/v2"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			defer pprof.StopCPUProfile()
		}
	}
	t := newTMPC()
	t.win.Resize(fyne.NewSize(800, 500))
	t.Run()
}
