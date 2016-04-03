package compiler

import "testing"

func TestMath(t *testing.T) {
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
