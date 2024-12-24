package screen_test

import (
	"testing"

	"github.com/shrimp332/snake/screen"
)

func TestParseSeqArrow(t *testing.T) {
	cases := map[string]screen.ScreenEvent{
		"\x1b[A": screen.ArrowUp,
		"\x1b[B": screen.ArrowDown,
		"\x1b[C": screen.ArrowRight,
		"\x1b[D": screen.ArrowLeft,
	}

	for c, event := range cases {
		e, err := screen.ParseSeq([]byte(c))
		if err != nil {
			t.Error(c, err)
			continue
		}

		if e.E != event {
			t.Error(c, e.E, event)
			continue
		}
	}
}

func TestParseSeqNewSize(t *testing.T) {
	cases := map[string]screen.ScreenEvent{
		"\x1b[1234;1234R": screen.NewSize,
	}

	for c, event := range cases {
		e, err := screen.ParseSeq([]byte(c))
		if err != nil {
			t.Error(c, err)
			continue
		}

		if e.E != event {
			t.Error(c, e.E, event)
			continue
		}

		p := screen.Position{Y: 1234, X: 1234}

		if p != *e.Pos() {
			t.Error(p, *e.Pos())
			continue
		}
	}
}
