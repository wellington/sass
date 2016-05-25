package compiler

import "testing"

func TestControl_inline_if(t *testing.T) {
	in := `div {
  blah: if($x, ("red.png", blah), "blue.png");
}`
	e := `div {
  blah: "red.png", blah; }
`
	runParse(t, in, e)
}

func TestControl_interp_if(t *testing.T) {
	in := `$file-1x: "budge.png";

@function fudge($str) {
  @return "assets/fudge/" + $str;
}

div {
  blah: if($x, fudge("#{$file-1x}"), "#{$file-1x}");
}`
	e := `div {
  blah: "assets/fudge/budge.png"; }
`
	runParse(t, in, e)
}
