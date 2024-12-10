package main

import (
	"fmt"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

const (
	asciiDown  = "\x1b[B"
	asciiUp    = "\x1b[A"
	asciiRight = "\x1b[C"
	asciiLeft  = "\x1b[D"
)

const (
	arrowKey   = 27
	arrowUp    = 65
	arrowDown  = 66
	arrowRight = 67
	arrowLeft  = 68
)

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

type Position struct {
	x, y int
}

type Snake struct {
	Head    Position
	Tail    []Position
	dir     Direction
	prevDir Direction
	length  int // score
	ate     bool
}

func (s *Snake) isTail(p Position) bool {
	for _, v := range s.Tail {
		if v.x == p.x && v.y == p.y {
			return true
		}
	}
	return false
}

type food struct {
	pos []Position
}

func (f *food) add(p Position) {
	f.pos = append(f.pos, p)
}

func (f *food) isFood(p Position) bool {
	for _, v := range f.pos {
		if v.x == p.x && v.y == p.y {
			return true
		}
	}
	return false
}

func (f *food) remove(p Position) {
	i := slices.IndexFunc(
		f.pos,
		func(pos Position) bool { return pos.x == p.x && pos.y == p.y },
	)
	f.pos[i] = f.pos[len(f.pos)-1]
	f.pos = f.pos[:len(f.pos)-1]
}

const maxFood = 3

const (
	snakeHead = "O"
	snakeBody = "o"
	foodIcon  = "x"
)

var screenSize Position

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	screenSize = getTSize()

	s := Snake{
		dir:     Right,
		prevDir: Right,
		length:  0,
	}
	s.Head.x = screenSize.x / 2
	s.Head.y = screenSize.y / 2

	f := food{}

	c := make(chan rune, 1)
	go func() {
		for {
			b := make([]byte, 1)
			_, err := os.Stdin.Read(b)
			if err != nil {
				panic(err)
			}
			c <- rune(b[0])
		}
	}()

	go func() {
		counter := 0
		for {
			counter++
			if len(f.pos) < maxFood && counter%5 == 0 {
        counter = 0
				fmt.Print("\x1b[s")
				foodx := rand.Intn(screenSize.x)
				foody := rand.Intn(screenSize.y)
				p := Position{
					x: foodx,
					y: foody,
				}
				f.add(p)
				moveCursorPos(p)
				fmt.Print(foodIcon)
				fmt.Print("\x1b[u")
			}

			if s.length > 0 {
				fmt.Print("\x1b[s")
				moveCursorPos(s.Tail[len(s.Tail)-1])
				fmt.Print(" ")
				fmt.Print("\x1b[u")
				s.Tail = s.Tail[:len(s.Tail)-1]
			}

			s.move()

			if s.isTail(s.Head) {
				screenPrintTop("Game Over, Final Score: ", s.length)
				term.Restore(int(os.Stdin.Fd()), oldState)
				os.Exit(0)
			}

			if f.isFood(s.Head) {
				s.length++
				s.Tail = append(s.Tail, s.Head)
				f.remove(s.Head)
			}
			screenPrint("Score: ", s.length)
			time.Sleep(300 * time.Millisecond)
		}
	}()

	for {
		r := <-c
		switch r {
		case 3: // Ctrl+C
			screenPrintTop("Game Over, Final Score: ", s.length)
			return
		case arrowKey:
			_ = <-c // ignore next byte
			r = <-c
			switch r {
			case arrowUp:
				if s.prevDir != Down {
					s.dir = Up
				}
			case arrowDown:
				if s.prevDir != Up {
					s.dir = Down
				}
			case arrowLeft:
				if s.prevDir != Right {
					s.dir = Left
				}
			case arrowRight:
				if s.prevDir != Left {
					s.dir = Right
				}
			}
		}
	}
}

func screenPrint(v ...any) {
	fmt.Print("\x1b[s")
	defer fmt.Print("\x1b[u")
	moveCursorPos(Position{screenSize.x + 1, 0})
	fmt.Print("\x1b[2K")
	fmt.Print(v...)
}

func screenPrintTop(v ...any) {
	fmt.Print("\x1b[s")
	defer fmt.Print("\x1b[u")
	moveCursorPos(Position{0, 0})
	fmt.Print("\x1b[2K")
	fmt.Print(v...)
}

func moveCursorPos(p Position) {
	fmt.Printf("\x1b[%d;%dH", p.x, p.y)
}

func getCursorPos() Position {
	fmt.Print("\x1b[6n")

	buf := []byte{}
	for {
		b := make([]byte, 1)
		_, err := os.Stdin.Read(b)
		if err != nil {
			panic(err)
		}
		if b[0] == 'R' {
			break
		} else {
			buf = append(buf, b[0])
		}

	}
	sPos := strings.Split(string(buf[2:]), ";")
	pos := Position{
		x: unwrap(strconv.Atoi(sPos[0])),
		y: unwrap(strconv.Atoi(sPos[1])),
	}

	return pos
}

/*
* Must be ran before Stdnin loop starts
* add mutex if needed to run later
 */
func getTSize() Position {
	fmt.Print("\x1b[s")
	moveCursorPos(Position{10000, 10000})
	p := getCursorPos()
	p.x-- // add space for screenPrint
	fmt.Print("\x1b[u")
	return p
}

/*
* dangerous, panics if err
 */
func unwrap[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func (s *Snake) move() {
	s.prevDir = s.dir
	moveCursorPos(s.Head)
	if s.length > 0 {
		fmt.Print(snakeBody)
		s.Tail = append([]Position{s.Head}, s.Tail...)
	} else {
		fmt.Print(" ")
	}

	switch s.dir {
	case Up:
		s.Head.x--
		if s.Head.x <= 0 {
			s.Head.x = screenSize.x
		}
	case Down:
		s.Head.x++
		if s.Head.x >= screenSize.x {
			s.Head.x = 0
		}
	case Right:
		s.Head.y++
		if s.Head.y >= screenSize.y {
			s.Head.y = 1
		}
	case Left:
		s.Head.y--
		if s.Head.y <= 1 {
			s.Head.y = screenSize.y
		}
	}

	moveCursorPos(s.Head)
	fmt.Print(snakeHead)
	moveCursorPos(s.Head)
}
