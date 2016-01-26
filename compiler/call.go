package compiler

import (
	"errors"
	"fmt"
	"image/color"
	"log"
	"strconv"

	"github.com/wellington/sass/ast"
)

var ErrNotFound = errors.New("function does not exist")

func init() {
	Register("rgb", rgb)
}

type callFunc func(args []ast.Expr) (*ast.BasicLit, error)

var funcs map[string]callFunc = make(map[string]callFunc)

func Register(name string, fn callFunc) {
	if _, ok := funcs[name]; ok {
		panic("already registered: " + name)
	}

	funcs[name] = fn
}

// This might not be enough
func evaluateCall(expr *ast.CallExpr) (*ast.BasicLit, error) {

	ident := expr.Fun.(*ast.Ident)

	fn, ok := funcs[ident.Name]
	if !ok {
		return nil, ErrNotFound
	}

	return fn(expr.Args)
}

// Builtin functions

func rgb(args []ast.Expr) (*ast.BasicLit, error) {
	if len(args) != 3 {
		return nil,
			fmt.Errorf("invalid number of args received expected 3")
	}

	lits := make([]*ast.BasicLit, 3)
	for i := range args {
		switch v := args[i].(type) {
		case *ast.BasicLit:
			lits[i] = v
		case *ast.KeyValueExpr:
			// Named argument parsing
			key := v.Key.(*ast.BasicLit)
			val := v.Value.(*ast.BasicLit)
			switch key.Value {
			case "$red":
				lits[0] = val
			case "$green":
				lits[1] = val
			case "$blue":
				lits[2] = val
			default:
				log.Fatal("unsupported", key.Value)
			}
		}

	}
	r, _ := strconv.Atoi(lits[0].Value)
	g, _ := strconv.Atoi(lits[1].Value)
	b, _ := strconv.Atoi(lits[2].Value)
	lit := ast.BasicLitFromColor(color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
	})
	// There's some stupidity in the color stuff, do a lookup
	// manually
	lit.Value = ast.LookupColor(lit.Value)
	return lit, nil
}
