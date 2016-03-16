package builtin

import "github.com/wellington/sass/ast"

type CallHandler func(expr *ast.CallExpr, args ...*ast.BasicLit) (*ast.BasicLit, error)

var reg func(s string, ch CallHandler, c CallHandle)

var chs = map[string]CallHandler{}

func BindRegister(old func(s string, ch CallHandler, c CallHandle)) {
	reg = old
	for k, v := range chs {
		reg(k, v, nil)
		delete(chs, k)
	}
	for k, v := range cs {
		reg(k, nil, v)
		delete(cs, k)
	}
}

func Register(s string, ch CallHandler) {
	if reg != nil {
		reg(s, ch, nil)
		return
	}
	chs[s] = ch
}

var cs = map[string]CallHandle{}

// CallHandle pass in Expr get out Expr. This replaces
// the limited CallHandler which can't work on map or lists
type CallHandle func(expr *ast.CallExpr, args ...ast.Expr) (ast.Expr, error)

// Reg registers a CallHandle for use by parser
func Reg(s string, ch CallHandle) {
	if reg != nil {
		reg(s, nil, ch)
		return
	}
	cs[s] = ch
}
