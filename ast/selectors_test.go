package ast

import (
	"testing"

	"github.com/wellington/sass/token"
)

func TestLevelTwo(t *testing.T) {
	r := selMultiply(" ", "div", "p")
	if e := "div p"; e != r {
		t.Errorf("got: %q wanted: %q", r, e)
	}

	r = selMultiply(" ", "div", "p, a")
	if e := "div p, div a"; e != r {
		t.Errorf("got: %q wanted: %q", r, e)
	}

	r = selMultiply(" ", "div", "p", "a, b")
	if e := "div p a, div p b"; e != r {
		t.Errorf("got: %q wanted: %q", r, e)
	}

	r = selMultiply(" ", "div", "&")
	if e := "div"; e != r {
		t.Errorf("got: %q wanted: %q", r, e)
	}

	r = selMultiply(" ", "c", "& + &")
	if e := "c + c"; e != r {
		t.Errorf("got: %q wanted: %q", r, e)
	}

	r = selMultiply(" ", "c", "&, &")
	if e := "c, c"; e != r {
		t.Errorf("got: %q wanted: %q", r, e)
	}

	r = selMultiply(" ", "a, b", "&")
	if e := "a, b"; e != r {
		t.Errorf("got: %q wanted: %q", r, e)
	}

	r = selMultiply(" ", "a, b", "&")
	if e := "a, b"; e != r {
		t.Errorf("got: %q wanted: %q", r, e)
	}

	r = selMultiply(" ", "c, d", "e, f")
	if e := "c e, c f, d e, d f"; e != r {
		t.Errorf("got: %q wanted: %q", r, e)
	}

	// TODO: better backreference support
	// r = selMultiply(" ", "a, b", "& + &")
	// if e := "a + a, b + a, a + b, b + b"; e != r {
	// 	 t.Errorf("got: %q wanted: %q", r, e)
	// }

}

type elt struct {
	tok token.Token
	lit string
	pos token.Pos
}

func TestSelExpand(t *testing.T) {
	s := "a + b > c &"
	parts := selExpand(s, 10, "")

	var elts = []elt{
		{token.STRING, "a ", 0},
		{token.ADD, "+", 2},
		{token.STRING, " b ", 3},
		{token.GTR, ">", 6},
		{token.STRING, " c ", 7},
		{token.AND, "&", 10},
	}

	for i := range parts {
		if parts[i].Kind != elts[i].tok {
			t.Errorf("token mismatch got: %s wanted: %s",
				parts[i].Kind, elts[i].tok)
		}

		if parts[i].ValuePos != elts[i].pos {
			t.Errorf("token mismatch got: %s wanted: %s",
				parts[i].ValuePos, elts[i].pos)
		}

		if parts[i].Value != elts[i].lit {
			t.Errorf("token mismatch got: %s wanted: %s",
				parts[i].Value, elts[i].lit)
		}
	}

}
