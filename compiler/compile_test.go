package compiler

import (
	"io/ioutil"
	"testing"
)

func TestRun(t *testing.T) {
	ctx := &Context{}
	ctx.Init()
	out, err := ctx.Run("../sass-spec/spec/basic/01_simple_css/input.scss")
	if err != nil {
		t.Fatal(err)
	}
	_ = out
	bs, err := ioutil.ReadFile("../sass-spec/spec/basic/01_simple_css/expected_output.css")
	if err != nil {
		t.Fatal(err)
	}

	if e := string(bs); e != out {
		t.Errorf("got:\n%s\nwanted:\n%s", out, e)
	}
}
