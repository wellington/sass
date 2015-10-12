package main

import (
	"bytes"
	"testing"
)

func init() {
	yyErrorVerbose = true
}

func TestParserSimple(t *testing.T) {
	buf := bytes.Buffer{}
	out = &buf
	in := bytes.NewBufferString(`div { color: red; }`)

	lexer := NewDefault(in)
	e := yyParse(lexer)
	if e != 0 {
		t.Fatal("parser reported error")
	}

	if e := `div{color:red;}`; e != buf.String() {
		t.Errorf("got: %s\nwanted: %s\n", buf.String(), e)
	}
}
