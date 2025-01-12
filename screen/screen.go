package screen

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/term"

	"github.com/shrimp332/snake/util"
)

var ErrNotSeq error = errors.New("Not a valid escape sequence")

type ScreenEvent byte

const (
	NewSize ScreenEvent = iota
	Char    ScreenEvent = iota

	ArrowUp    ScreenEvent = 'A'
	ArrowDown  ScreenEvent = 'B'
	ArrowRight ScreenEvent = 'C'
	ArrowLeft  ScreenEvent = 'D'
)

const (
	Escape = '\x1b'
	Csi    = '['
)

type Event struct {
	E ScreenEvent
	s string
	r *rune
	p *Position // might be nil, dependant on ScreenEvent
}

func (e *Event) Rune() *rune {
	return e.r
}

func (e *Event) Pos() *Position {
	return e.p
}

func (e *Event) Seq() string {
	return e.s
}

type Position struct{ X, Y int }

type Screen struct {
	size     *Position
	Debug    bool
	oldState *term.State
	raw      bool
	Q        chan *Event
}

func (s *Screen) Raw() bool {
	return s.raw
}

func (s *Screen) SetSize(p Position) {
	p.Y-- // Make room for status bar
	if s.Debug {
		p.Y--
	}
	s.size = &p
}

func (s *Screen) Size() *Position {
	return s.size
}

func (s *Screen) PrintStatus(v ...any) {
	fmt.Print("\x1b[40m")
	defer fmt.Print("\x1b[49m")
	s.PrintAt(Position{Y: s.size.Y + 1, X: 1}, "\x1b[2K")
	s.PrintAt(Position{Y: s.size.Y + 1, X: 1}, v...)
}

func (s *Screen) PrintDebug(v ...any) {
	if !s.Debug {
		return
	}
	fmt.Print("\x1b[40m")
	defer fmt.Print("\x1b[49m")
	s.PrintAt(Position{Y: s.size.Y + 2, X: 1}, "\x1b[2K")
	s.PrintAt(Position{Y: s.size.Y + 2, X: 1}, v...)
}

func (s *Screen) Start() {
	fmt.Print("\x1b[?1049h") // alternate
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	fmt.Print("\x1b[?25l") // hide cursor
	s.Clear()
	s.raw = true
	s.oldState = oldState
	s.Q = make(chan *Event)
	go s.readIn()
}

func (s *Screen) Clear() {
	fmt.Print("\x1b[H\x1b[2J")
}

func (s *Screen) readIn() {
	for {
		b := make([]byte, 12)
		_, err := os.Stdin.Read(b)
		if err != nil {
			panic(err)
		}
		e, err := ParseSeq(b)
		if err != nil {
			if errors.Is(ErrNotSeq, err) {
				ru := rune(b[0])
				e = &Event{
					E: Char,
					r: &ru,
				}
			} else {
				panic(err)
			}
		}
		s.Q <- e
	}
}

func (s *Screen) Cleanup() {
	if !s.raw {
		return
	}
	if err := term.Restore(int(os.Stdin.Fd()), s.oldState); err != nil {
		panic(err)
	}
	s.raw = false
	s.oldState = nil
	fmt.Print("\x1b[?25h")   // Reset cursor
	fmt.Print("\x1b[?1049l") // alternate
}

// Results in NewSize Event to Screen.Q
func (s *Screen) GetSize() {
	if !s.raw {
		return
	}
	s.moveCursor(Position{9999, 9999})
	fmt.Print("\x1b[6n")
}

func (s *Screen) moveCursor(p Position) {
	if !s.raw {
		return
	}
	fmt.Printf("\x1b[%d;%dH", p.Y, p.X)
}

var mu sync.Mutex

func (s *Screen) PrintAt(p Position, v ...any) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Print("\x1b[s")
	defer fmt.Print("\x1b[u")
	s.moveCursor(p)
	fmt.Print(v...)
}

func ParseSeq(b []byte) (*Event, error) {
	if b[0] != Escape || b[1] != Csi {
		return nil, ErrNotSeq
	}

	e := Event{
		s: string(b),
	}

	if b[2] >= byte(ArrowUp) && b[2] <= byte(ArrowLeft) {
		e.E = ScreenEvent(b[2])
		return &e, nil
	}

	if b[2] <= 48 && b[2] >= 57 {
		return nil, ErrNotSeq
	}

	i := slices.IndexFunc(
		b,
		func(b byte) bool { return b == byte('R') },
	)

	if b[i] != 'R' {
		return nil, ErrNotSeq
	}

	sPos := strings.Split(string(b[2:i]), ";")
	pos := &Position{
		Y: util.Unwrap(strconv.Atoi(sPos[0])),
		X: util.Unwrap(strconv.Atoi(sPos[1])),
	}
	e.E = NewSize
	e.p = pos

	return &e, nil
}
