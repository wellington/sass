package ast

import "fmt"

// ToIdent converts expressions to Ident
func ToIdent(expr Expr) *Ident {
	switch v := expr.(type) {
	case *BasicLit:
		return &Ident{
			Name:    v.Value,
			NamePos: v.ValuePos,
		}
	case *Ident:
		return v
	default:
		fmt.Printf("Failed to cast expr to Ident % #v\n", v)
	}
	return nil
}
