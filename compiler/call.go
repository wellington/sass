package compiler

import (
	"errors"
	"image/color"
	"log"
	"strconv"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/token"
)

var ErrNotFound = errors.New("function does not exist")

func init() {
	Register("rgb", rgb)
	Register("rgba", rgba)
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

func parseColors(args []ast.Expr) (color.RGBA, error) {
	lits := make([]*ast.BasicLit, 4)
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
			case "$alpha":
				lits[3] = val
			default:
				log.Fatal("unsupported", key.Value)
			}
		}

	}
	var err error
	ints := make([]uint8, 4)
	if lits[3] != nil && err == nil {
		var f float64
		f, err = strconv.ParseFloat(lits[3].Value, 32)
		ints[3] = uint8(f * 100)
	}

	if lits[0] != nil && lits[0].Kind == token.COLOR {
		c := ast.ColorFromHexString(lits[0].Value)
		c.A = ints[3]
		return c, nil
	}

	for i := range lits[:3] {
		if lits[i] != nil {
			var n int
			n, err = strconv.Atoi(lits[i].Value)
			if err != nil {
				break
			}
			ints[i] = uint8(n)
		}
	}
	return color.RGBA{
		R: ints[0],
		G: ints[1],
		B: ints[2],
		A: ints[3],
	}, err
}

func rgb(args []ast.Expr) (*ast.BasicLit, error) {
	c, err := parseColors(args)
	if err != nil {
		return nil, err
	}
	lit := ast.BasicLitFromColor(c)
	// There's some stupidity in the color stuff, do a lookup
	// manually
	lit.Value = ast.LookupColor(lit.Value)
	return lit, nil
}

func rgba(args []ast.Expr) (*ast.BasicLit, error) {
	c, err := parseColors(args)
	if err != nil {
		return nil, err
	}
	lit := ast.BasicLitFromColor(c)
	// There's some stupidity in the color stuff, do a lookup
	// manually
	lit.Value = ast.LookupColor(lit.Value)
	return lit, nil
}
