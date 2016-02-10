package ast

import (
	"encoding/hex"
	"fmt"
	"image/color"
	"log"
	"strconv"
	"unicode/utf8"

	"github.com/wellington/sass/token"
)

const maxUint8 = ^uint8(0)

// Simplify matching color hashes to CSS names
// https://developer.mozilla.org/en-US/docs/Web/CSS/color_value
var cssColors = map[string]string{
	"#000000": "black",
	"#c0c0c0": "silver",
	"#808080": "gray",
	"#ffffff": "white",
	"#800000": "maroon",
	"#ff0000": "red",
	"#800080": "purple",
	"#ff00ff": "magenta",
	"#008000": "green",
	"#00ff00": "lime",
	"#808000": "olive",
	"#ffff00": "yellow",
	"#000080": "navy",
	"#0000ff": "blue",
	"#008080": "teal",
	"#00ffff": "cyan",
	"#ffa500": "orange",
	"#f0f8ff": "aliceblue",
	"#faebd7": "antiquewhite",
	"#7fffd4": "aquamarine",
	"#f0ffff": "azure",
	"#f5f5dc": "beige",
	"#ffe4c4": "bisque",
	"#ffebcd": "blanchedalmond",
	"#8a2be2": "blueviolet",
	"#a52a2a": "brown",
	"#deb887": "burlywood",
	"#5f9ea0": "cadetblue",
	"#7fff00": "chartreuse",
	"#d2691e": "chocolate",
	"#ff7f50": "coral",
	"#6495ed": "cornflowerblue",
	"#fff8dc": "cornsilk",
	"#dc143c": "crimson",
	"#00008b": "darkblue",
	"#008b8b": "darkcyan",
	"#b8860b": "darkgoldenrod",
	"#a9a9a9": "darkgray",
	"#006400": "darkgreen",
	// "#a9a9a9": "darkgrey",
	"#bdb76b": "darkkhaki",
	"#8b008b": "darkmagenta",
	"#556b2f": "darkolivegreen",
	"#ff8c00": "darkorange",
	"#9932cc": "darkorchid",
	"#8b0000": "darkred",
	"#e9967a": "darksalmon",
	"#8fbc8f": "darkseagreen",
	"#483d8b": "darkslateblue",
	"#2f4f4f": "darkslategray",
	"#00ced1": "darkturquoise",
	"#9400d3": "darkviolet",
	"#ff1493": "deeppink",
	"#00bfff": "deepskyblue",
	"#696969": "dimgray",
	"#1e90ff": "dodgerblue",
	"#b22222": "firebrick",
	"#fffaf0": "floralwhite",
	"#228b22": "forestgreen",
	"#dcdcdc": "gainsboro",
	"#f8f8ff": "ghostwhite",
	"#ffd700": "gold",
	"#daa520": "goldenrod",
	"#adff2f": "greenyellow",
	"#f0fff0": "honeydew",
	"#ff69b4": "hotpink",
	"#cd5c5c": "indianred",
	"#4b0082": "indigo",
	"#fffff0": "ivory",
	"#f0e68c": "khaki",
	"#e6e6fa": "lavender",
	"#fff0f5": "lavenderblush",
	"#7cfc00": "lawngreen",
	"#fffacd": "lemonchiffon",
	"#add8e6": "lightblue",
	"#f08080": "lightcoral",
	"#e0ffff": "lightcyan",
	"#fafad2": "lightgoldenrodyellow",
	"#d3d3d3": "lightgray",
	"#90ee90": "lightgreen",
	"#ffb6c1": "lightpink",
	"#ffa07a": "lightsalmon",
	"#20b2aa": "lightseagreen",
	"#87cefa": "lightskyblue",
	"#778899": "lightslategray",
	"#b0c4de": "lightsteelblue",
	"#ffffe0": "lightyellow",
	"#32cd32": "limegreen",
	"#faf0e6": "linen",
	"#66cdaa": "mediumaquamarine",
	"#0000cd": "mediumblue",
	"#ba55d3": "mediumorchid",
	"#9370db": "mediumpurple",
	"#3cb371": "mediumseagreen",
	"#7b68ee": "mediumslateblue",
	"#00fa9a": "mediumspringgreen",
	"#48d1cc": "mediumturquoise",
	"#c71585": "mediumvioletred",
	"#191970": "midnightblue",
	"#f5fffa": "mintcream",
	"#ffe4e1": "mistyrose",
	"#ffe4b5": "moccasin",
	"#ffdead": "navajowhite",
	"#fdf5e6": "oldlace",
	"#6b8e23": "olivedrab",
	"#ff4500": "orangered",
	"#da70d6": "orchid",
	"#eee8aa": "palegoldenrod",
	"#98fb98": "palegreen",
	"#afeeee": "paleturquoise",
	"#db7093": "palevioletred",
	"#ffefd5": "papayawhip",
	"#ffdab9": "peachpuff",
	"#cd853f": "peru",
	"#ffc0cb": "pink",
	"#dda0dd": "plum",
	"#b0e0e6": "powderblue",
	"#bc8f8f": "rosybrown",
	"#4169e1": "royalblue",
	"#8b4513": "saddlebrown",
	"#fa8072": "salmon",
	"#f4a460": "sandybrown",
	"#2e8b57": "seagreen",
	"#fff5ee": "seashell",
	"#a0522d": "sienna",
	"#87ceeb": "skyblue",
	"#6a5acd": "slateblue",
	"#708090": "slategray",
	"#fffafa": "snow",
	"#00ff7f": "springgreen",
	"#4682b4": "steelblue",
	"#d2b48c": "tan",
	"#d8bfd8": "thistle",
	"#ff6347": "tomato",
	"#40e0d0": "turquoise",
	"#ee82ee": "violet",
	"#f5deb3": "wheat",
	"#f5f5f5": "whitesmoke",
	"#9acd32": "yellowgreen",
	"#663399": "rebeccapurple",
}

// LookupColor finds a CSS name for a hex, if available. Otherwise,
// it returns the hex representation.
func LookupColor(s string) string {
	// check for CSS color name
	if name, ok := cssColors[s]; ok {
		fmt.Println(s, name)
		return name
	}
	return s
}

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

func ColorFromHexString(s string) color.RGBA {
	return colorFromHexString(s)
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
		// Shittttttt..... need better internal
		// representation of colors
		s := string(in)
		var found bool
		for key, color := range cssColors {
			if s == color {
				found = true
				in = []byte(key)[1:]
			}
		}
		if !found {
			panic("Invalid color hex: " + string(in))
		}
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
	return "#" + hex.EncodeToString([]byte{uint8(r)}) +
		hex.EncodeToString([]byte{uint8(g)}) +
		hex.EncodeToString([]byte{uint8(b)})
}

func BasicLitFromColor(c color.Color) *BasicLit {
	return &BasicLit{
		Kind:  token.COLOR,
		Value: colorToHex(c),
	}
}

func colorOpColor(tok token.Token, x *BasicLit, y *BasicLit) (*BasicLit, error) {
	colX := colorFromHexString(x.Value)
	colY := colorFromHexString(y.Value)

	var z color.RGBA
	z.R = overflowMath(tok, colX.R, colY.R)
	z.G = overflowMath(tok, colX.G, colY.G)
	z.B = overflowMath(tok, colX.B, colY.B)

	s := colorToHex(z)
	return &BasicLit{
		Kind:  token.COLOR,
		Value: LookupColor(s),
	}, nil
}

// Sass does not allow overflow uint8
func overflowMath(tok token.Token, a, b uint8) uint8 {
	var c uint

	switch tok {
	case token.ADD:
		c = uint(a) + uint(b)
		if c > uint(maxUint8) {
			return maxUint8
		}
	case token.SUB:
		c = uint(uint8(a) - uint8(b))
		// If result is greater than a, then overflow happend
		if c > uint(a) {
			return 0
		}
	case token.QUO:
		c = uint(a / b)
	}

	return uint8(c)
}

func colorOpInt(tok token.Token, c *BasicLit, i *BasicLit) (*BasicLit, error) {
	col := colorFromHexString(c.Value)
	j, err := strconv.Atoi(i.Value)
	if err != nil {
		return nil, err
	}

	col.R = overflowMath(tok, col.R, uint8(j))
	col.G = overflowMath(tok, col.G, uint8(j))
	col.B = overflowMath(tok, col.B, uint8(j))

	s := colorToHex(col)

	return &BasicLit{
		Kind:  token.COLOR,
		Value: LookupColor(s),
		// Created Expr doesn't have a position
	}, err
}
