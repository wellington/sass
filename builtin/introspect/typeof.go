package introspect

import (
	"errors"
	"fmt"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/builtin"
	"github.com/wellington/sass/token"
)

func init() {
	builtin.Register("inspect($value)", inspect)
	builtin.Register("unit($number)", unit)
	builtin.Register("type-of($value)", typeOf)
}

func unit(call *ast.CallExpr, args ...*ast.BasicLit) (*ast.BasicLit, error) {
	in := *args[0]
	lit := &ast.BasicLit{
		Kind:     token.QSTRING,
		ValuePos: call.Pos(),
	}
	switch in.Kind {
	case token.UEM:
		lit.Value = "em"
	case token.UPX:
		lit.Value = "px"
	case token.UPCT:
		lit.Value = "%"
	case token.INT, token.FLOAT:
		lit.Value = ""
	case token.STRING, token.QSTRING, token.QSSTRING:
		return nil, fmt.Errorf(`$number: "%s" is not a number for unit`, in.Value)
	default:
		return nil, errors.New("unsupported type for type-of")
	}
	lit.Value = lit.Value
	return lit, nil
}

func inspect(call *ast.CallExpr, args ...*ast.BasicLit) (*ast.BasicLit, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments (%d for 1) for 'inspect'", len(args))
	}
	return args[0], nil
}

func typeOf(call *ast.CallExpr, args ...*ast.BasicLit) (*ast.BasicLit, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments (%d for 1) for 'type-of'", len(args))
	}
	lit := *args[0]
	lit.Kind = token.STRING
	switch args[0].Kind {
	case token.COLOR:
		lit.Value = "color"
	case token.INT, token.FLOAT:
		lit.Value = "number"
	case token.STRING, token.QSSTRING, token.QSTRING:
		lit.Value = "string"
	default:
		lit.Kind = token.ILLEGAL
	}
	return &lit, nil
}
