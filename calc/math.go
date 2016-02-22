package calc

import (
	"log"
	"strconv"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/token"
)

// Resolve simple math to create a basic lit
func Resolve(in ast.Expr) *ast.BasicLit {
	return resolve(in)
}

func resolve(in ast.Expr) *ast.BasicLit {
	switch v := in.(type) {
	case *ast.UnaryExpr:
		return v.X.(*ast.BasicLit)
	case *ast.BinaryExpr:
		return binary(v)
	case *ast.BasicLit:
		return v
	default:
		log.Fatalf("unsupported calc.resolve % #v\n", v)
	}
	return nil
}

// binary takes a BinaryExpr and simplifies it to a
func binary(in *ast.BinaryExpr) *ast.BasicLit {
	left, right := resolve(in.X), resolve(in.Y)
	out := &ast.BasicLit{
		ValuePos: left.Pos(),
		Kind:     token.STRING,
	}
	switch in.Op {
	case token.ADD:
		if left.Kind == token.INT && left.Kind == left.Kind {
			l, _ := strconv.Atoi(left.Value)
			r, _ := strconv.Atoi(right.Value)
			out.Kind = token.INT
			out.Value = strconv.Itoa(l + r)
		} else {
			out.Value = left.Value + right.Value
		}
	default:
		log.Fatalf("unsupported: %s", in.Op)
	}
	return out
}
