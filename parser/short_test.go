package parser

import "testing"

var valids = []string{
	"$color: red;",
	`$color: "black" !global;`,
	"p15a: 10 - #a2B;",
	"p18: 10 #a2b + 1;",
	"p20: rgb(10,10,10) + #010001;",
	"@mixin foo($a: one, $b) { $x: inside $a; } inner { @include foo(); @include foo(two); }",
	"rgb(255, $blue: 0, $green: 255);",
	"mix(rgba(#f0e, $alpha: .5)+#111, #00f);",
	"$a: h#{ello + world};",
	"a#{id} { a: b; }",
	"b: type-of(12#{3});",
	"a+b,c{ d e, f~g, > i {}}",
	"div { /*comment*/ }",
	// Easy way to test resolveCall
	"div { a: #{rgb(4,5,6)+1}; }",
}

func TestValid(t *testing.T) {
	for _, src := range valids {
		checkErrors(t, src, src)
	}
}

var invalids = []string{
	"mix(#111);",
}
