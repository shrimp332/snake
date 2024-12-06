package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

var (
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

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	fmt.Print("\x1b[s")
  defer fmt.Print("\x1b[u")

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

	for {
		r := <-c
		switch r {
		case 3: // Ctrl+C
      return
		case arrowKey:
			_ = <-c // ignore next byte
			r = <-c
			switch r {
			case arrowUp:
				fmt.Print(asciiUp)
			case arrowDown:
				fmt.Print(asciiDown)
			case arrowLeft:
				fmt.Print(asciiLeft)
			case arrowRight:
				fmt.Print(asciiRight)
			}
		default: // debug
			fmt.Print(r)
			fmt.Print(",")
		}
	}
}

func unwrap[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
