package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func spec(t *testing.T, path ...string) (io.Reader, string) {
	p := append([]string{"..", "sass-spec", "spec"}, path...)
	dir := filepath.Join(p...)
	bs, err := ioutil.ReadFile(filepath.Join(dir, "input.scss"))
	if err != nil {
		t.Fatal(err)
	}
	bbs, err := ioutil.ReadFile(filepath.Join(dir, "expected.compact.css"))
	if err != nil {
		t.Fatal(err)
	}
	return bytes.NewBuffer(bs), string(bbs)
}

func runner(t *testing.T, path string) (bytes.Buffer, string) {
	in, e := spec(t, path)
	buf, err := setupParser(t, in)
	if err != nil {
		t.Fatal(path, err)
	}
	return buf, e
}

func TestSassSpec_basic(t *testing.T) {
	var e string
	var buf bytes.Buffer
	yyDebug = 0

	buf, e = runner(t, "basic/00_empty")
	if e != buf.String() {
		t.Errorf("got: %s\nwanted: %s\n", buf.String(), e)
	}

	buf, e = runner(t, "basic/01_simple_css")
	if e != buf.String() {
		t.Errorf("got: %s\nwanted: %s\n", buf.String(), e)
	}

	buf, e = runner(t, "basic/02_simple_nesting")
	if e != buf.String() {
		t.Errorf("got: %s\nwanted: %s\n", buf.String(), e)
	}

	buf, e = runner(t, "basic/03_simple_variable")
	if e != buf.String() {
		t.Errorf("got: %s\nwanted: %s\n", buf.String(), e)
	}

}

func TestTest(t *testing.T) {
	yyDebug = 5
	debug = true
	buf, e := runner(t, "basic/04_basic_variables")
	if e != buf.String() {
		t.Errorf("got: %s\nwanted: %s\n", buf.String(), e)
	}

}
