package main

import (
	"math/rand"
	"os"
	"slices"
	"time"

	"github.com/shrimp332/snake/screen"
)

var debugMode bool

func init() {
	debug := os.Getenv("S_DEBUG")
	if debug == "true" || debug == "1" {
		debugMode = true
	}
}

const maxFood = 3

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

const (
	snakeHead = "O"
	snakeBody = "o"
	foodIcon  = "\x1b[31mx\x1b[0m" // red x
)

func main() {
	scr := screen.Screen{}
	scr.Start()
	defer scr.Cleanup()

	s := Snake{
		dir:     Right,
		prevDir: Right,
		length:  0,
		delay:   500 * time.Millisecond,
	}
	go eventHandler(&scr, &s)
	scr.GetSize()

	for scr.Size == nil {
	}
	s.Head.Y = scr.Size.Y / 2
	s.Head.X = scr.Size.X / 2
	f := food{}

	counter := 0
	for {
		counter++
		if len(f.pos) < maxFood && counter%5 == 0 {
			counter = 0
			for !f.add(&s, &scr) {
			}
			scr.PrintAt(foodIcon, f.pos[len(f.pos)-1])
		}

		if s.length > 0 {
			scr.PrintAt(" ", s.Tail[len(s.Tail)-1])
			s.Tail = s.Tail[:len(s.Tail)-1]
		}

		s.move(&scr)

		if s.isTail(s.Head) {
			// screenPrintTop("Game Over, Final Score: ", s.length)
			scr.Cleanup()
			os.Exit(0)
		}

		if f.isFood(s.Head) {
			s.length++
			s.Tail = append(s.Tail, s.Head)
			f.remove(s.Head)
			if s.delay > 100*time.Millisecond {
				s.delay -= 25 * time.Millisecond
			}
		}
		if debugMode {
			// screenPrintDebug("Head: ", s.Head, " Delay: ", s.delay, " Food: ", f.pos)
		}
		// screenPrint("Score: ", s.length)
		time.Sleep(s.delay)
	}
}

func eventHandler(scr *screen.Screen, s *Snake) {
	for {
		e := <-scr.Q
		switch e.E {
		case screen.NewSize:
			scr.Size = e.Pos()
		case screen.ArrowUp:
			if s.prevDir != Down {
				s.dir = Up
			}
		case screen.ArrowDown:
			if s.prevDir != Up {
				s.dir = Down
			}
		case screen.ArrowRight:
			if s.prevDir != Left {
				s.dir = Right
			}
		case screen.ArrowLeft:
			if s.prevDir != Right {
				s.dir = Left
			}
		case screen.Char:
			if *e.Rune() == 3 || *e.Rune() == 4 {
				scr.Cleanup()
				os.Exit(0)
			}
		}
	}
}

type Snake struct {
	Head    screen.Position
	Tail    []screen.Position
	dir     Direction
	prevDir Direction
	length  int // score
	ate     bool
	delay   time.Duration
}

func (s *Snake) isTail(p screen.Position) bool {
	for _, v := range s.Tail {
		if v == p {
			return true
		}
	}
	return false
}

func (s *Snake) move(scr *screen.Screen) {
	s.prevDir = s.dir
	if s.length > 0 {
		scr.PrintAt(snakeBody, s.Head)
		s.Tail = append([]screen.Position{s.Head}, s.Tail...)
	} else {
		scr.PrintAt(" ", s.Head)
	}
	switch s.dir {
	case Up:
		s.Head.Y--
		if s.Head.Y <= 0 {
			s.Head.Y = scr.Size.Y
		}
	case Down:
		s.Head.Y++
		if s.Head.Y > scr.Size.Y {
			s.Head.Y = 1
		}
	case Right:
		s.Head.X++
		if s.Head.X > scr.Size.X {
			s.Head.X = 1
		}
	case Left:
		s.Head.X--
		if s.Head.X <= 0 {
			s.Head.X = scr.Size.X
		}
	}

	scr.PrintAt(snakeHead, s.Head)
}

type food struct {
	pos []screen.Position
}

func (f *food) add(s *Snake, scr *screen.Screen) bool {
	foodx := rand.Intn(scr.Size.X - 1)
	foody := rand.Intn(scr.Size.Y - 1)
	foodx++
	foody++
	p := screen.Position{
		X: foodx,
		Y: foody,
	}
	if f.isFood(p) || s.Head == p || s.isTail(p) {
		return false
	}
	f.pos = append(f.pos, p)
	return true
}

func (f *food) isFood(p screen.Position) bool {
	for _, v := range f.pos {
		if p == v {
			return true
		}
	}
	return false
}

func (f *food) remove(p screen.Position) {
	i := slices.IndexFunc(
		f.pos,
		func(pos screen.Position) bool { return p == pos },
	)
	f.pos[i] = f.pos[len(f.pos)-1]
	f.pos = f.pos[:len(f.pos)-1]
}
