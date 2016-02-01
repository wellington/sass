package compiler

import (
	"errors"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/compiler/builtin"
)

var ErrNotFound = errors.New("function does not exist")

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
		return notfoundCall(expr), nil
	}
	return fn(expr.Args)
}

// there's no such thing as a failure in Sass. Resolve idents in callexpr
// and return result as BasicLit
func notfoundCall(call *ast.CallExpr) (lit *ast.BasicLit) {

	return
}

func init() {
	Register("rgb", builtin.RGB)
	Register("rgba", builtin.RGBA)
}
