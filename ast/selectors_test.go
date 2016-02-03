package ast

import "testing"

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
