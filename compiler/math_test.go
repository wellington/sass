package compiler

import "testing"

func TestMath_unit_convert(t *testing.T) {
	in := `
div {
  w: 4px + w;
  o: 3px + 3px + 3px;
  p: 4 + 1px;
  no: 15 / 3 / 5;
  yes: ( 15 / 3 / 5 );
}
`
	e := `div {
  w: 4pxw;
  o: 9px;
  p: 5px;
  no: 15/3/5;
  yes: 1; }
`
	runParse(t, in, e)
}

func TestMath_fractions(t *testing.T) {
	in := `
div {
  a: 1 + 2;
  b: 3 + 3/4;
  c: 1/2 + 1/2;
  d: 1/2;
}
`
	e := `div {
  a: 3;
  b: 3.75;
  c: 1;
  d: 1/2; }
`
	runParse(t, in, e)
}

func TestMath_list(t *testing.T) {
	in := `
div {
  e: 1 + (5/10 4 7 8);
  f: (5/10 2 3) + 1;
  g: (15 / 3) / 5;
}
`
	e := `div {
  e: 15/10 4 7 8;
  f: 5/10 2 31;
  g: 1; }
`
	runParse(t, in, e)
}

func TestMath_var(t *testing.T) {
	in := `
$three: 3;
div {
  k: 15 / $three;
  l: 15 / 5 / $three;
}
`
	e := `div {
  k: 5;
  l: 1; }
`
	runParse(t, in, e)
}

func TestMath_mixed_unit(t *testing.T) {

	in := `
div {
  r: 16em * 4;
  s: (5em / 2);
  t: 5em/2;
}
`
	e := `div {
  r: 64em;
  s: 2.5em;
  t: 5em/2; }
`
	runParse(t, in, e)
}
