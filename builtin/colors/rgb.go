package colors

import (
	"fmt"
	"image/color"
	"log"
	"strconv"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/builtin"
	"github.com/wellington/sass/token"
)

func init() {
	builtin.Register("rgb($red:0, $green:0, $blue:0)", rgb)
	builtin.Register("rgba($red:0, $green:0, $blue:0, $alpha:0)", rgba)
	builtin.Register("mix($color1:0, $color2:0, $weight:0.5)", mix)
	builtin.Register("invert($color)", invert)
}

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

func rgb(args []*ast.BasicLit) (*ast.BasicLit, error) {
	log.Printf("rgb args: red: %s green: %s blue: %s\n",
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

func rgba(args []*ast.BasicLit) (*ast.BasicLit, error) {
	log.Printf("rgba args: red: %s green: %s blue: %s alpha: %s\n",
		args[0].Value, args[1].Value, args[2].Value, args[3].Value)

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

func mix(args []*ast.BasicLit) (*ast.BasicLit, error) {
	fmt.Printf("mix:\narg0: % #v\narg1: % #v\narg2: % #v\n",
		args[0], args[1], args[2])
	// parse that weight
	wt, err := strconv.ParseFloat(args[2].Value, 8)
	// Parse percentage ie. 50%
	if err != nil {
		var i float64
		fmt.Println("parsing", args[2].Value)
		_, err := fmt.Sscanf(args[2].Value, "%f%%", &i)
		if err != nil {
			log.Fatal(err)
		}
		wt = i / 100
	}
	c1 := ast.ColorFromHexString(args[0].Value)
	c2 := ast.ColorFromHexString(args[1].Value)
	var r, g, b, a float64
	r = wt*float64(c1.R) + (1-wt)*float64(c2.R)
	g = wt*float64(c1.G) + (1-wt)*float64(c2.G)
	b = wt*float64(c1.B) + (1-wt)*float64(c2.B)
	a = wt*float64(c1.A) + (1-wt)*float64(c2.A)
	fmt.Println("r", r, "g", g, "b", b, "a", a)
	ret := color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: uint8(a),
	}
	lit := ast.BasicLitFromColor(ret)
	fmt.Printf("mix return: % #v\n", lit.Value)
	return lit, nil
}

func invert(args []*ast.BasicLit) (*ast.BasicLit, error) {
	c := ast.ColorFromHexString(args[0].Value)

	c.R = 255 - c.R
	c.G = 255 - c.G
	c.B = 255 - c.B

	return ast.BasicLitFromColor(c), nil
}
