package unit

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/token"
)

// Unit represents the CSS unit being described
type Unit int

// Why does this exist?
//
// var unitTypes = map[string]string{
// 	"in":   "distance",
// 	"cm":   "distance",
// 	"pc":   "distance",
// 	"mm":   "distance",
// 	"pt":   "distance",
// 	"px":   "distance",
// 	"deg":  "angle",
// 	"grad": "angle",
// 	"rad":  "angle",
// 	"turn": "angle",
// }

const (
	// INVALID unit can not be converted
	INVALID Unit = iota
	IN
	CM
	MM
	PC
	PX
	PT
	DEG
	GRAD
	RAD
	TURN
	// NOUNIT represents float or int that perform ops but are not
	// tied to a unit
	NOUNIT
)

func (u Unit) String() string {
	switch u {
	case IN:
		return "IN"
	case CM:
		return "CM"
	case MM:
		return "MM"
	case PC:
		return "PC"
	case PX:
		return "PX"
	case PT:
		return "PT"
	case DEG:
		return "DEG"
	case GRAD:
		return "GRAD"
	case RAD:
		return "RAD"
	case TURN:
		return "TURN"
	case NOUNIT:
		return "NOUNIT"
	}
	return "invalid"
}

var mlook = map[token.Token]Unit{
	token.FLOAT: NOUNIT,
	token.INT:   NOUNIT,
	token.UIN:   IN,
	token.UCM:   CM,
	token.UMM:   MM,
	token.UPC:   PC,
	token.UPX:   PX,
	token.UPT:   PT,
	token.DEG:   DEG,
	token.GRAD:  GRAD,
	token.RAD:   RAD,
	token.TURN:  TURN,
}

func unitLookup(tok token.Token) Unit {
	u, ok := mlook[tok]
	if !ok {
		return INVALID
	}
	return u
}

func tokLookup(u Unit) token.Token {
	for t, unit := range mlook {
		if unit == u {
			return t
		}
	}
	return token.ILLEGAL
}

var cnv = [...]float64{
	PC: 6,
	CM: 2.54,
	MM: 25.4,
	PT: 72,
	PX: 96,
}

var unitconv = [...][11]float64{
	IN: {
		IN:   1,
		CM:   2.54,
		PC:   6,
		MM:   25.4,
		PT:   72,
		PX:   96,
		DEG:  1,
		GRAD: 1,
		RAD:  1,
		TURN: 1,
	},
	PC: {
		IN:   1.0 / cnv[PC],
		CM:   2.54 / cnv[PC],
		PC:   6.0 / cnv[PC],
		MM:   25.4 / cnv[PC],
		PT:   72 / cnv[PC],
		PX:   96 / cnv[PC],
		DEG:  1,
		GRAD: 1,
		RAD:  1,
		TURN: 1,
	},
	CM: {
		IN:   1 / cnv[CM],
		CM:   2.54 / cnv[CM],
		PC:   6 / cnv[CM],
		MM:   25.4 / cnv[CM],
		PT:   72 / cnv[CM],
		PX:   96 / cnv[CM],
		DEG:  1,
		GRAD: 1,
		RAD:  1,
		TURN: 1,
	},
	MM: {
		IN:   1 / cnv[MM],
		CM:   2.54 / cnv[MM],
		PC:   6 / cnv[MM],
		MM:   25.4 / cnv[MM],
		PT:   72 / cnv[MM],
		PX:   96 / cnv[MM],
		DEG:  1,
		GRAD: 1,
		RAD:  1,
		TURN: 1,
	},
	PT: {
		IN:   1 / cnv[PT],
		CM:   2.54 / cnv[PT],
		PC:   6 / cnv[PT],
		MM:   25.4 / cnv[PT],
		PT:   72 / cnv[PT],
		PX:   96 / cnv[PT],
		DEG:  1,
		GRAD: 1,
		RAD:  1,
		TURN: 1,
	},
	PX: {
		IN:   1 / cnv[PX],
		CM:   2.54 / cnv[PX],
		PC:   6 / cnv[PX],
		MM:   25.4 / cnv[PX],
		PT:   72 / cnv[PX],
		PX:   96 / cnv[PX],
		DEG:  1,
		GRAD: 1,
		RAD:  1,
		TURN: 1,
	},
	// conversion not useful for these
	DEG: {
		IN:   1,
		CM:   1,
		PC:   1,
		MM:   1,
		PT:   1,
		PX:   1,
		DEG:  1,
		GRAD: 4 / 3.6,
		RAD:  math.Pi / 180.0,
		TURN: 1.0 / 360.0,
	},
	GRAD: {
		IN:   1,
		CM:   1,
		PC:   1,
		MM:   1,
		PT:   1,
		PX:   1,
		DEG:  3.6 / 4,
		GRAD: 1,
		RAD:  math.Pi / 200.0,
		TURN: 1.0 / 400.0,
	},
	RAD: {
		IN:   1,
		CM:   1,
		PC:   1,
		MM:   1,
		PT:   1,
		PX:   1,
		DEG:  180 / math.Pi,
		GRAD: 200 / math.Pi,
		RAD:  1,
		TURN: math.Pi / 2,
	},
	TURN: {
		IN:   1,
		CM:   1,
		PC:   1,
		MM:   1,
		PT:   1,
		PX:   1,
		DEG:  360,
		GRAD: 400,
		RAD:  2.0 * math.Pi,
		TURN: 1,
	},
	NOUNIT: {
		IN:   1,
		CM:   1,
		PC:   1,
		MM:   1,
		PT:   1,
		PX:   1,
		DEG:  1,
		GRAD: 1,
		RAD:  1,
		TURN: 1,
	},
}

func init() {
	ast.RegisterKind(Combine, token.UPX, token.INT)
}

// Combine lit with specified kind rules
func Combine(op token.Token, x, y *ast.BasicLit) (*ast.BasicLit, error) {
	// So we have some non-standard units, convert to INT/FLOAT
	// and send to another handler. Return always matches the unit
	// of x
	// var unitx, unity token.Token

	m, err := NewNum(x)
	if err != nil {
		return nil, err
	}
	n, err := NewNum(y)
	if err != nil {
		return nil, err
	}

	m.Op(op, m, n)
	fmt.Printf("<<< % #v\n", m)
	return m.Lit()
}

// Num represents a float with units. This isn't useful for
// unitless operations, since "calc" already does this.
type Num struct {
	pos token.Pos
	f   float64
	Unit
}

// NewNum initializes a Num from a BasicLit. Kind will hold the unit
// the number portion is always treated as a float.
func NewNum(lit *ast.BasicLit) (*Num, error) {
	val := lit.Value
	// TODO: scanner should remove unit
	kind := lit.Kind
	val = strings.TrimSuffix(lit.Value, token.Tokens[kind])
	f, err := strconv.ParseFloat(val, 64)
	return &Num{f: f, Unit: unitLookup(kind)}, err
}

func (z *Num) String() string {
	return strconv.FormatFloat(z.f, 'G', -1, 64) + tokLookup(z.Unit).String()
}

// Lit attempts to convert Num back into a Lit.
func (z *Num) Lit() (*ast.BasicLit, error) {
	return &ast.BasicLit{
		Kind:     tokLookup(z.Unit),
		Value:    z.String(),
		ValuePos: z.pos,
	}, nil
}

// Convert src to z, applying proper conversion to src
func (z *Num) Convert(src *Num) *Num {
	u := z.Unit
	var cv float64
	n := &Num{}
	if u == NOUNIT {
		// nounit inherits unit of src
		u = src.Unit
		cv = 1
	} else {
		cv = unitconv[z.Unit][u]
	}

	n.f = src.f * cv
	n.Unit = u
	return n
}

// Op returns the sum of x and y using the specified Op
func (z *Num) Op(op token.Token, x, y *Num) *Num {
	switch op {
	case token.ADD:
		return z.Add(x, y)
	case token.MUL:
		return z.Mul(x, y)
	case token.QUO:
		return z.Quo(x, y)
	case token.SUB:
		return z.Sub(x, y)
	default:
		panic(fmt.Errorf("unsupported unit op: %s", op))
	}
}

// Add returns the sum of x and y
func (z *Num) Add(x, y *Num) *Num {
	// n controls output unit
	fmt.Printf("adding % #v + % #v\n", x, y)
	fmt.Println(x, y, z)
	z.Convert(y)
	a, b := z.Convert(x), z.Convert(y)
	fmt.Println("conver", a, b)
	fmt.Println(z)
	z.f = a.f + b.f
	return z
}

// Sub returns the subtraction of x and y
func (z *Num) Sub(x, y *Num) *Num {
	// n controls output unit
	a, b := x.Convert(z), y.Convert(z)
	z.f = a.f - b.f
	return z
}

// Mul returns the multiplication of x and y
func (z *Num) Mul(x, y *Num) *Num {
	// n controls output unit
	a, b := x.Convert(z), y.Convert(z)
	z.f = a.f * b.f
	return z
}

// Quo returns the division of x and y
func (z *Num) Quo(x, y *Num) *Num {
	// n controls output unit
	a, b := x.Convert(z), y.Convert(z)
	z.f = a.f / b.f
	return z
}
