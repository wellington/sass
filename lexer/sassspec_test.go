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

func TestSassSpec_basic(t *testing.T) {
	in, e := spec(t, "basic/01_simple_css")

	buf, err := setupParser(t, in)
	if err != nil {
		t.Fatal(err)
	}

	if e != buf.String() {
		t.Errorf("got: %s\nwanted: %s\n", buf.String(), e)
	}
}
