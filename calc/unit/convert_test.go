package unit

import "testing"

var samp = []*Num{
	&Num{Unit: IN, f: 1},
	&Num{Unit: CM, f: 2.54},
	&Num{Unit: MM, f: 25.4},
	&Num{Unit: PT, f: 72},
	&Num{Unit: PX, f: 96},
}

func copy(n *Num) *Num {
	x := *n
	return &x
}

func TestConvert(t *testing.T) {

	for i := range samp {
		x := copy(samp[0])
		x.Convert(samp[i])
		if x.Unit != samp[0].Unit {
			t.Errorf("got: %s wanted: %s", samp[0].Unit, x.Unit)
		}
		if e := samp[i].f; e != x.f {
			t.Errorf("got: %f wanted: %f", x.f, e)
		}
	}
}

func TestAdd(t *testing.T) {
	for i := range samp {
		x := copy(samp[0])
		x.Add(samp[0], samp[i])
		if x.Unit != samp[0].Unit {
			t.Errorf("got: %s wanted: %s", samp[0].Unit, x.Unit)
		}
		if e := 2.0; e != x.f {
			t.Errorf("got: %f wanted: %f", x.f, e)
		}
	}
}

func TestSub(t *testing.T) {
	for i := range samp {
		x := copy(samp[0])
		x.Sub(samp[0], samp[i])
		if x.Unit != samp[0].Unit {
			t.Errorf("got: %s wanted: %s", samp[0].Unit, x.Unit)
		}
		if e := 0.0; e != x.f {
			t.Errorf("got: %f wanted: %f", x.f, e)
		}
	}
}

func TestMul(t *testing.T) {
	for i := range samp {
		x := copy(samp[0])
		x.Mul(samp[0], samp[i])
		if x.Unit != samp[0].Unit {
			t.Errorf("got: %s wanted: %s", samp[0].Unit, x.Unit)
		}
		if e := 1.0; e != x.f {
			t.Errorf("got: %f wanted: %f", x.f, e)
		}
	}
}

func TestQuo(t *testing.T) {
	for i := range samp {
		x := copy(samp[0])
		x.Quo(samp[0], samp[i])
		if x.Unit != samp[0].Unit {
			t.Errorf("got: %s wanted: %s", samp[0].Unit, x.Unit)
		}
		if e := 1.0; e != x.f {
			t.Errorf("got: %f wanted: %f", x.f, e)
		}
	}
}
