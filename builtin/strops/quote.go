package strops

import (
	"strconv"
	"strings"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/builtin"
	"github.com/wellington/sass/token"
)

func init() {
	builtin.Register("unquote($string)", unquote)
}

func unquote(call *ast.CallExpr, args ...*ast.BasicLit) (*ast.BasicLit, error) {
	in := *args[0]
	lit := &ast.BasicLit{
		Kind:     token.STRING,
		ValuePos: in.ValuePos,
		Value:    Unquote(in.Value),
	}
	// Because in Ruby Sass, there is no failure though libSass fails
	// very easily
	return lit, nil
}

const (
	sassEscape = `\`
	goEscape   = `\u`
	quote      = `"`
)

func Unquote(in string) string {
	return unescape(in)
}

// unquote converts Sass's bizarre unicode escape format to valid
// unicode text
func unescape(in string) string {
	ss := strings.Split(in, sassEscape)
	// No sass unicode
	if len(ss) == 1 {
		return in
	}
	// Attempt unquote on each Sass escape found
	for i, s := range ss {
		if len(s) == 0 {
			continue
		}
		in := quote + goEscape + s + quote
		unq, err := strconv.Unquote(in)
		// if unquote was successful, replace
		if err == nil {
			ss[i] = unq
		}
	}

	return strings.Join(ss, "")
}
