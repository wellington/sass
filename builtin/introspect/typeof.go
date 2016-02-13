package introspect

import (
	"errors"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/builtin"
	"github.com/wellington/sass/token"
)

func init() {
	builtin.Register("type-of($value)", typeOf)
}

func typeOf(call *ast.CallExpr, args ...*ast.BasicLit) (*ast.BasicLit, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments (2 for 1) for 'type-of'")
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
