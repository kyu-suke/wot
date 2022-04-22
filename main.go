package main

import (
	"flag"

	"github.com/kyu-suke/wot/gms"
	"github.com/nsf/termbox-go"
)

func main() {
	var s = flag.Int("s", 100, "speed")
	var h = flag.Int("h", 25, "stage height")
	var w = flag.Int("w", 80, "stage width")
	var g = flag.String("g", "snake", "game")
	flag.Parse()

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	switch *g {
	case "snake":
		gms.StartSnakeGame(*s, *h, *w)
	default:
		gms.StartSnakeGame(*s, *h, *w)
	}
}
