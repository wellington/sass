package main

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func init() {
	yyErrorVerbose = true
}

func setupParser(t *testing.T, in io.Reader) (bytes.Buffer, error) {
	buf := bytes.Buffer{}
	out = &buf
	lexer := NewDefault(in)
	e := yyParse(lexer)
	if e != 0 {
		return buf, errors.New("parser reported error")
	}
	return buf, nil
}

func TestParserSimple(t *testing.T) {
	// yyDebug = 10
	in := bytes.NewBufferString(`div { color: red; }`)
	buf, err := setupParser(t, in)
	if err != nil {
		t.Fatal(err)
	}

	if e := `div{color:red;}`; e != buf.String() {
		t.Errorf("got: %s\nwanted: %s\n", buf.String(), e)
	}
}

func TestParserProps(t *testing.T) {
	in := bytes.NewBufferString(`p{color: blue; background-color: red;}`)
	buf, err := setupParser(t, in)
	if err != nil {
		t.Fatal(err)
	}

	if e := `p{color:blue;background-color:red;}`; e != buf.String() {
		t.Errorf("got: %s\nwanted: %s\n", buf.String(), e)
	}
}

func TestParserNested(t *testing.T) {
	in := bytes.NewBufferString(`div { p { color: red; } }`)
	buf, err := setupParser(t, in)
	if err != nil {
		t.Fatal(err)
	}

	if e := `div p{color:red;}`; e != buf.String() {
		t.Errorf("got: %s\nwanted: %s\n", buf.String(), e)
	}
}

func TestParserDoubleNestception(t *testing.T) {
	in := bytes.NewBufferString(`div { div {p {color: red; } } }`)
	buf, err := setupParser(t, in)
	if err != nil {
		t.Fatal(err)
	}

	if e := `div div p{color:red;}`; e != buf.String() {
		t.Errorf("got: %s\nwanted: %s\n", buf.String(), e)
	}

}
