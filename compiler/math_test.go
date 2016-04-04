package compiler

import "testing"

func TestMath_simple(t *testing.T) {
	//o: 3px + 3px + 3px;
	//p: 4 + 1px;
	//not: 3 / 3;
	in := `
div {
  o: 3px + 3px + 3px;
  p: 4 + 1px;
  no: 15 / 3 / 5;
  yes: ( 15 / 3 / 5 );
}
`
	e := `div {
  o: 9px;
  p: 5px;
  no: 15/3/5;
  yes: 1; }
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
