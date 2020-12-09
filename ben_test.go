package main

import (
	"testing"

	"uk.ac.bris.cs/gameoflife/gol"
)

func BenchmarkGol(b *testing.B) {
	p := gol.Params{
		Turns:       1000,
		Threads:     2,
		ImageWidth:  512,
		ImageHeight: 512,
	}
	//alive := readAliveCounts(p.ImageWidth, p.ImageHeight)
	events := make(chan gol.Event)
	gol.Run(p, events, nil)

	//var cells []util.Cell

	for event := range events {
		switch event.(type) {
		case gol.FinalTurnComplete:
			//cells = e.Alive
		}
	}
	//cells = append(cells, util.Cell{0, 0})
}
