package introspect

import (
	"testing"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/token"
)

func TestTypeOf(t *testing.T) {
	call := &ast.CallExpr{}
	lit, err := typeOf(call, &ast.BasicLit{
		Kind:  token.INT,
		Value: "1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if e := "number"; lit.Value != e {
		t.Errorf("got: %s wanted: %s", lit.Value, e)
	}

	lit, err = typeOf(call, &ast.BasicLit{
		Kind:  token.STRING,
		Value: "a",
	})
	if err != nil {
		t.Fatal(err)
	}
	if e := "string"; lit.Value != e {
		t.Errorf("got: %s wanted: %s", lit.Value, e)
	}
}
