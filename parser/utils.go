package parser

import (
	"fmt"
	"strings"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/token"
)

// unwrapQuotes performs conversions from quoted strings to regular strings
// "word" => word
// 'word' => word
func unwrapQuotes(in *ast.BasicLit) *ast.BasicLit {
	switch in.Kind {
	case token.QSTRING:
		in.Kind = token.STRING
		in.Value = strings.Trim(in.Value, `"`)
	case token.QSSTRING:
		in.Kind = token.STRING
		in.Value = strings.Trim(in.Value, `'`)
	}
	if token.STRING != in.Kind {
		panic(fmt.Errorf("invalid type %s: % #v\n", in.Kind, in))
	}
	return in
}
