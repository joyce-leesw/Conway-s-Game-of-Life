package gol

import (
	"fmt"

	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events    chan<- Event
	ioCommand chan<- ioCommand
	ioIdle    <-chan bool

	filename chan<- string
	outputQ  chan<- uint8 // this was diffrent before outputQ    chan<- uint8
	inputQ   <-chan uint8
}

// type cell struct {
// 	X, Y int
// }

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.
	turn := 0

	initialWorld := make([][]byte, p.ImageHeight)
	for i := 0; i < (p.ImageHeight); i++ {
		initialWorld[i] = make([]byte, p.ImageWidth)
	}

	//open the pgm file to game of life
	fileName := fmt.Sprintf("%vx%v", p.ImageWidth, p.ImageHeight)

	//read the game of life and convert the pgm into a slice of slices
	//initialWorld := readPgmImage(p, fileName
	c.ioCommand <- ioInput
	c.filename <- fileName

	for i := 0; i < (p.ImageHeight); i++ {
		for j := 0; j < (p.ImageHeight); j++ {
			initialWorld[i][j] = <-c.inputQ
		}
	}

	// TODO: For all initially alive cells send a CellFlipped Event.
	//calculateAliveCells(p, initialWorld)
	for _, cellQ := range calculateAliveCells(p, initialWorld) {
		c.events <- CellFlipped{0, cellQ}
	}

	// TODO: Execute all turns of the Game of Life.
	var state State
	world := initialWorld
	for turnf := 0; turnf < p.Turns; turnf++ {
		world = calculateNextState(p, world)

		for _, cellQ := range calculateAliveCells(p, world) {
			c.events <- CellFlipped{turnf, cellQ}
		}
		c.events <- AliveCellsCount{turnf, len(calculateAliveCells(p, world))}
		c.events <- TurnComplete{turnf}
		c.events <- StateChange{turnf, state}
	}

	// TODO: Send correct Events when required, e.g. CellFlipped, TurnComplete and FinalTurnComplete.
	//		 See event.go for a list of all events.
	turn = p.Turns


	c.events <- FinalTurnComplete{turn, calculateAliveCells(p, world)}
	c.ioCommand <- ioOutput
	c.filename <- fileName
	
	for i := 0; i < (p.ImageHeight); i++ {
		for j := 0; j < (p.ImageHeight); j++ {
			c.outputQ <- world[i][j]
		}
	}

	c.events <- ImageOutputComplete{turn, fileName}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

func calculateAliveCells(p Params, world [][]byte) []util.Cell {
	aliveCells := []util.Cell{}
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == 255 {
				aliveCells = append(aliveCells, util.Cell{X: x, Y: y})
			}
		}
	}
	return aliveCells
}

func calculateNextState(p Params, world [][]byte) [][]byte {

	var counter, nextX, lastX, nextY, lastY int //can I use a byte here instead
	newWS := make([][]byte, p.ImageHeight)
	for i := 0; i < len(world); i++ {
		newWS[i] = make([]byte, p.ImageWidth)
	}

	//fmt.Println(newWS)

	for y, s := range world {

		for x, sl := range s {

			counter = 0

			//Set the nextX and lastX variables
			if x == len(s)-1 { //are we looking at the last element of the slice
				nextX = 0
				lastX = x - 1
			} else if x == 0 { //are we looking at the first element of the slice
				nextX = x + 1
				lastX = len(s) - 1
			} else { //we are looking at any element that is not the first of last element of a slice
				nextX = x + 1
				lastX = x - 1
			}

			//Set the nextY and lastY variables
			if y == len(world)-1 { //are we looking at the last element of the slice
				nextY = 0
				lastY = y - 1
			} else if y == 0 { //are we looking at the first element of the slice
				nextY = y + 1
				lastY = len(world) - 1
			} else { //we are looking at any element that is not the first of last element of a slice
				nextY = y + 1
				lastY = y - 1
			}

			if 255 == s[nextX] {
				counter++
			} //look E
			if 255 == s[lastX] {
				counter++
			} //look W

			if 255 == world[lastY][lastX] {
				counter++
			} //look NW
			if 255 == world[lastY][x] {
				counter++
			} //look N
			if 255 == world[lastY][nextX] {
				counter++
			} //look NE

			if 255 == world[nextY][lastX] {
				counter++
			} //look SW
			if 255 == world[nextY][x] {
				counter++
			} //look S
			if 255 == world[nextY][nextX] {
				counter++
			} //look SE

			//Live cells
			if sl == 255 {
				if counter < 2 { //"any live cell with fewer than two live neighbours dies"
					newWS[y][x] = 0
				} else if counter > 3 { //"any live cell with more than three live neighbours dies"
					newWS[y][x] = 0
				} else { //"any live cell with two or three live neighbours is unaffected"
					newWS[y][x] = 255
				}

			}

			//Dead cell -- not MGS
			if sl == 0 {
				if counter == 3 { //"any dead cell with exactly three live neighbours becomes alive"
					newWS[y][x] = 255
				} else {
					newWS[y][x] = 0 // Dead cells elsewise stay dead
				}

			}

		}

	}
	world = newWS

	return world
}
