package calc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/token"
)

// Resolve simple math to create a basic lit
func Resolve(in ast.Expr) (*ast.BasicLit, error) {
	return resolve(in)
}

func resolve(in ast.Expr) (*ast.BasicLit, error) {
	x := &ast.BasicLit{
		ValuePos: in.Pos(),
	}
	var err error
	switch v := in.(type) {
	case *ast.StringExpr:
		list := make([]string, 0, len(v.List))
		for _, l := range v.List {
			lit, err := resolve(l)
			if err != nil {
				return nil, err
			}
			list = append(list, lit.Value)
		}
		x.Kind = v.Kind
		x.Value = strings.Join(list, "")
	case *ast.ListLit:
		// During expr simplification, list are just string
		delim := " "
		if v.Comma {
			delim = ", "
		}
		ss := make([]string, len(v.Value))
		for i := range v.Value {
			ss[i] = v.Value[i].(*ast.BasicLit).Value
		}
		return &ast.BasicLit{
			Value:    strings.Join(ss, delim),
			ValuePos: v.Pos(),
		}, nil
	case *ast.UnaryExpr:
		x = v.X.(*ast.BasicLit)
	case *ast.BinaryExpr:
		x, err = binary(v)
	case *ast.BasicLit:
		x = v
	case *ast.Ident:
		if v.Obj == nil {
			return nil, fmt.Errorf("calc: undefined variable %s", v.Name)
		}
		rhs := v.Obj.Decl.(*ast.AssignStmt).Rhs
		kind := token.INT
		var val []string
		for _, x := range rhs {
			lit, err := resolve(x)
			if err != nil {
				return nil, err
			}
			// TODO: insufficient!
			if lit.Kind != kind {
				kind = lit.Kind
			}
			val = append(val, lit.Value)
		}
		// TODO: commas are missing
		x.Value = strings.Join(val, ", ")
		x.Kind = kind
	default:
		err = fmt.Errorf("unsupported calc.resolve % #v\n", v)
		panic(err)
	}
	return x, err
}

// binary takes a BinaryExpr and simplifies it to a
func binary(in *ast.BinaryExpr) (*ast.BasicLit, error) {
	left, err := resolve(in.X)
	if err != nil {
		return nil, err
	}
	right, err := resolve(in.Y)
	if err != nil {
		return nil, err
	}
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
		err = fmt.Errorf("unsupported: %s", in.Op)
	}
	return out, err
}
