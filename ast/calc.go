package ast

import (
	"encoding/hex"
	"image/color"
	"log"
	"strconv"
	"unicode/utf8"

	"github.com/wellington/sass/token"
)

func (x *BasicLit) Op(tok token.Token, y *BasicLit) *BasicLit {
	if x.Kind == token.COLOR {
		if y.Kind == token.INT {
			z, err := colorOpInt(tok, x, y)
			if err != nil {
				log.Fatal(err)
			}
			return z
		} else if y.Kind == token.COLOR {
			z, err := colorOpColor(tok, x, y)
			if err != nil {
				log.Fatal(err)
			}
			return z
		}
	}
	return nil
}

func colorFromHexString(s string) color.RGBA {
	return colorFromHex([]byte(s))
}

func colorFromHex(in []byte) color.RGBA {

	pound, w := utf8.DecodeRune(in)
	if pound == '#' {
		in = in[w:]
	}

	if len(in) == 3 {
		in = []byte{in[0], in[0], in[1], in[1], in[2], in[2]}
	}

	if len(in) != 6 {
		panic("Invalid color hex: " + string(in))
	}

	r, g, b, a := in[0:2], in[2:4], in[4:6], []byte{255, 255}

	hex.Decode(r, r)
	hex.Decode(g, g)
	hex.Decode(b, b)

	return color.RGBA{
		R: r[0],
		G: g[0],
		B: b[0],
		A: a[0],
	}
}

func colorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return hex.EncodeToString([]byte{uint8(r)}) +
		hex.EncodeToString([]byte{uint8(g)}) +
		hex.EncodeToString([]byte{uint8(b)})
}

func colorOpColor(tok token.Token, x *BasicLit, y *BasicLit) (*BasicLit, error) {
	colX := colorFromHexString(x.Value)
	colY := colorFromHexString(y.Value)

	return &BasicLit{
		Kind: token.COLOR,
		Value: "#" + colorToHex(color.RGBA{
			R: colX.R + colY.R,
			G: colX.G + colY.G,
			B: colX.B + colY.B,
		}),
	}, nil
}

func colorOpInt(tok token.Token, c *BasicLit, i *BasicLit) (*BasicLit, error) {
	col := colorFromHexString(c.Value)
	j, err := strconv.Atoi(i.Value)
	if err != nil {
		return nil, err
	}
	switch tok {
	case token.ADD:
		col.R += uint8(j)
		col.G += uint8(j)
		col.B += uint8(j)
	case token.SUB:
		col.R -= uint8(j)
		col.G -= uint8(j)
		col.B -= uint8(j)
	}

	return &BasicLit{
		Kind:  token.COLOR,
		Value: "#" + colorToHex(col),
		// Created Expr doesn't have a position
	}, err
}
