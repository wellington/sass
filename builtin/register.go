package builtin

import "github.com/wellington/sass/ast"

type CallHandler func(args []*ast.BasicLit) (*ast.BasicLit, error)

var reg func(s string, ch CallHandler)

var chs = map[string]CallHandler{}

func BindRegister(fn func(s string, ch CallHandler)) {
	reg = fn
	for k, v := range chs {
		reg(k, v)
		delete(chs, k)
	}
}

func Register(s string, ch CallHandler) {
	if reg != nil {
		reg(s, ch)
	} else {
		chs[s] = ch
	}
}
