package compiler

import "testing"

func TestMath(t *testing.T) {
	in := `
$stuff: 1 2 3;

$three: 3;

div {
  o: 3px + 3px + 3px;
  p: 4 + 1px;
}
`
	e := `div {
  o: 9px;
  p: 5px; }
`
	runParse(t, in, e)
}
