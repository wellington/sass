package parser

import "testing"

var valids = []string{
	"$color: red;",
	`$color: "black";`,
	"p15a: 10 - #a2B;",
	"p18: 10 #a2b + 1;",
	"p20: rgb(10,10,10) + #010001;",
}

func TestValid(t *testing.T) {
	for _, src := range valids {
		checkErrors(t, src, src)
	}
}
