package main

import (
	"bytes"
	"io"
	"testing"
)

func init() {
	yyErrorVerbose = true
}

func setupParser(t *testing.T, in io.Reader) bytes.Buffer {
	buf := bytes.Buffer{}
	out = &buf
	lexer := NewDefault(in)
	e := yyParse(lexer)
	if e != 0 {
		t.Fatal("parser reported error")
	}
	return buf
}

func TestParserSimple(t *testing.T) {
	in := bytes.NewBufferString(`div { color: red; }`)
	buf := setupParser(t, in)

	if e := `div{color:red;}`; e != buf.String() {
		t.Errorf("got: %s\nwanted: %s\n", buf.String(), e)
	}
}

func TestParserNested(t *testing.T) {
	yyDebug = 10
	in := bytes.NewBufferString(`div { p { color: red; } }`)
	buf := setupParser(t, in)

	if e := `div p{color:red;}`; e != buf.String() {
		t.Errorf("got: %s\nwanted: %s\n", buf.String(), e)
	}
}
