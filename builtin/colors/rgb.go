package colors

import (
	"fmt"
	"image/color"
	"log"
	"strconv"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/token"
)

func resolveDecl(ident *ast.Ident) []*ast.BasicLit {
	var lits []*ast.BasicLit
	switch decl := ident.Obj.Decl.(type) {
	case *ast.AssignStmt:
		call := decl.Rhs[0].(*ast.CallExpr)
		for i := range call.Args {
			lits = append(lits, call.Args[i].(*ast.BasicLit))
		}
	case *ast.BasicLit:
		lits = append(lits, decl)
	default:
		log.Fatalf("can not resolve: % #v\n", decl)
	}
	return lits
}

func parseColors(args []*ast.BasicLit) (color.RGBA, error) {
	lits := make([]*ast.BasicLit, 0, 4)
	for i := range args {
		v := args[i]
		switch v.Kind {
		case token.VAR:
			log.Fatalf("VAR % #v\n", v)
		case token.FLOAT, token.INT:
		case token.COLOR:
			return ast.ColorFromHexString(v.Value), nil
		default:
			log.Fatalf("unsupported kind %s % #v\n", v.Kind, v)
		}
		lits = append(lits, v)
	}
	var err error
	ints := make([]uint8, 4)
	if len(lits) > 3 {
		if lits[3] != nil && err == nil {
			var f float64
			f, err = strconv.ParseFloat(lits[3].Value, 32)
			ints[3] = uint8(f * 100)
		}
	}

	if lits[0] != nil && lits[0].Kind == token.COLOR {
		c := ast.ColorFromHexString(lits[0].Value)
		c.A = ints[3]
		return c, nil
	}

	for i := range lits[:3] {
		if lits[i] != nil {
			var n int
			n, err = strconv.Atoi(lits[i].Value)
			if err != nil {
				break
			}
			ints[i] = uint8(n)
		}
	}
	ret := color.RGBA{
		R: ints[0],
		G: ints[1],
		B: ints[2],
		A: ints[3],
	}
	fmt.Printf("color % #v\n", ret)
	return ret, err
}

func RGB(args []*ast.BasicLit) (*ast.BasicLit, error) {
	fmt.Printf("red: %s green: %s blue: %s\n",
		args[0].Value, args[1].Value, args[2].Value)
	c, err := parseColors(args)
	if err != nil {
		return nil, err
	}
	lit := ast.BasicLitFromColor(c)
	// There's some stupidity in the color stuff, do a lookup
	// manually
	lit.Value = ast.LookupColor(lit.Value)
	return lit, nil
}

func RGBA(args []*ast.BasicLit) (*ast.BasicLit, error) {
	fmt.Println("rgba", args)
	c, err := parseColors(args)
	if err != nil {
		return nil, err
	}
	_ = c
	return nil, nil
}
