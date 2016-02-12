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
	builtin.Register("red($color)", red)
	builtin.Register("blue($color)", blue)
	builtin.Register("green($color)", green)
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
	ints := make([]uint8, 4)
	var ret color.RGBA
	var u uint8
	var pos int
	for i := range args {
		if pos < i {
			pos = i
		}
		v := args[i]
		switch v.Kind {
		case token.VAR:
			log.Fatalf("VAR % #v\n", v)
		case token.FLOAT:
			f, err := strconv.ParseFloat(args[i].Value, 8)
			if err != nil {
				return ret, err
			}
			// Has to be alpha, or bust
			u = uint8(f * 100)
		case token.INT:
			i, err := strconv.Atoi(v.Value)
			if err != nil {
				return ret, err
			}
			u = uint8(i)
		case token.COLOR:
			if i != 0 {
				return ret, fmt.Errorf("hex is only allowed as the first argumetn found: % #v", v)
			}
			c := ast.ColorFromHexString(v.Value)
			ret = c
			// This is only allowed as the first argument
			pos = pos + 3
		default:
			log.Fatalf("unsupported kind %s % #v\n", v.Kind, v)
		}
		ints[pos] = u
	}
	if ints[0] > 0 {
		ret.R = ints[0]
	}
	if ints[1] > 0 {
		ret.G = ints[1]
	}
	if ints[2] > 0 {
		ret.B = ints[2]
	}
	if ints[3] > 0 {
		ret.A = ints[3]
	}
	return ret, nil
}

func onecolor(which string, args []*ast.BasicLit) (*ast.BasicLit, error) {
	c, err := parseColors(args)
	if err != nil {
		return nil, err
	}
	lit := &ast.BasicLit{
		Kind: token.INT,
	}
	switch which {
	case "red":
		lit.Value = strconv.Itoa(int(c.R))
	case "green":
		lit.Value = strconv.Itoa(int(c.G))
	case "blue":
		lit.Value = strconv.Itoa(int(c.B))
	default:
		panic("not a onecolor")
	}
	return lit, nil
}

func red(call *ast.CallExpr, args ...*ast.BasicLit) (*ast.BasicLit, error) {
	return onecolor("red", args)
}

func green(call *ast.CallExpr, args ...*ast.BasicLit) (*ast.BasicLit, error) {
	return onecolor("green", args)
}

func blue(call *ast.CallExpr, args ...*ast.BasicLit) (*ast.BasicLit, error) {
	return onecolor("blue", args)
}

func rgb(call *ast.CallExpr, args ...*ast.BasicLit) (*ast.BasicLit, error) {
	log.Printf("rgb args: red: %s green: %s blue: %s\n",
		args[0].Value, args[1].Value, args[2].Value)
	c, err := parseColors(args)
	if err != nil {
		return nil, err
	}

	return colorOutput(c, &ast.BasicLit{}), nil
}

func rgba(call *ast.CallExpr, args ...*ast.BasicLit) (*ast.BasicLit, error) {
	fmt.Println("rgba args:", args)
	log.Printf("rgba args: red: %s green: %s blue: %s alpha: %s\n",
		args[0].Value, args[1].Value, args[2].Value, args[3].Value)

	c, err := parseColors(args)
	if err != nil {
		return nil, err
	}
	fmt.Printf("c: % #v\n", c)

	return colorOutput(c, call), nil
}

func mix(call *ast.CallExpr, args ...*ast.BasicLit) (*ast.BasicLit, error) {
	fmt.Printf("mix:\narg0: % #v\narg1: % #v\narg2: % #v\n",
		args[0], args[1], args[2])
	// parse that weight
	wt, err := strconv.ParseFloat(args[2].Value, 8)
	// Parse percentage ie. 50%
	if err != nil {
		var i float64
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
	ret := color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: uint8(a),
	}

	return colorOutput(ret, call.Args[0]), nil
}

func invert(call *ast.CallExpr, args ...*ast.BasicLit) (*ast.BasicLit, error) {
	c := ast.ColorFromHexString(args[0].Value)

	c.R = 255 - c.R
	c.G = 255 - c.G
	c.B = 255 - c.B

	return colorOutput(c, call.Args[0]), nil
}

// colorOutput inspects the context to determine the appropriate output
func colorOutput(c color.RGBA, outTyp ast.Expr) *ast.BasicLit {
	ctx1 := outTyp
	lit := &ast.BasicLit{
		Kind: token.COLOR,
	}
	fmt.Printf("output % #v\n", ctx1)
	switch ctx := ctx1.(type) {
	case *ast.CallExpr:
		switch ctx.Fun.(*ast.Ident).Name {
		case "rgb":
			lit.Value = fmt.Sprintf("%s(%d, %d, %d)",
				"rgb", c.R, c.G, c.B,
			)
		case "rgba":
			i := int(c.A) * 10000
			f := float32(i) / 1000000
			lit.Value = fmt.Sprintf("%s(%d, %d, %d, %.g)",
				"rgba", c.R, c.G, c.B, f,
			)
		default:
			log.Fatal("unsupported ident", ctx.Fun.(*ast.Ident).Name)
		}
	case *ast.BasicLit:
		lit = ast.BasicLitFromColor(c)
	}
	return lit
}
