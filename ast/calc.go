package ast

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/wellington/sass/token"
)

// ErrNoCombine indicates that combination is not necessary
// The result should be treated as a sasslist or other form
// and input whitespace may be essential to compiler output.
var ErrNoCombine = errors.New("no op to perform")

// ErrIllegalOp indicate parsing errors on operands
var ErrIllegalOp = errors.New("operand is illegal")

type kind struct {
	unit    token.Token
	combine func(op token.Token, x, y *BasicLit, combine bool) (*BasicLit, error)
}

var kinds []kind

// RegisterKind enables additional external Operations. These could
// be color math or other non-literal math unsupported directly
// in ast.
func RegisterKind(fn func(op token.Token, x, y *BasicLit, combine bool) (*BasicLit, error), units ...token.Token) {
	for _, unit := range units {
		kinds = append(kinds, kind{unit: unit, combine: fn})
	}
}

// Op processes x op y applying unit conversions as necessary.
// combine forces operations on unitless numbers. By default,
// unitless numbers are not combined.
func Op(op token.Token, x, y *BasicLit, combine bool) (*BasicLit, error) {
	defer func() {
		fmt.Printf("kind: %s op: %s y: %s combine: %t x: % #v y: % #v\n",
			x.Kind, op, y.Kind, combine, x, y)
	}()
	if x.Kind == token.ILLEGAL || y.Kind == token.ILLEGAL {
		return nil, ErrIllegalOp
	}

	switch op {
	case token.MUL, token.ADD:
		// always combine these
		combine = true
	}

	kind := x.Kind
	var fn func(token.Token, *BasicLit, *BasicLit, bool) (*BasicLit, error)
	switch {
	case kind == token.COLOR:
		fn = colorOp
	case kind == token.INT:
		switch y.Kind {
		case token.INT:
			fn = intOp
		case token.STRING:
			fn = stringOp
		case token.FLOAT:
			fn = floatOp
		}
	case kind == token.FLOAT:
		switch y.Kind {
		case token.STRING:
			fn = stringOp
		default:
			fn = floatOp
		}
	case kind == token.STRING || x.Kind != y.Kind:
		fmt.Println("string op?")
		fn = stringOp
	}

	// math operations do not happen unless explicity enforced
	// on * and /
	if op == token.QUO || op == token.MUL {
		if fn != nil && !combine {
			fn = stringOp
		}
	}

	// no functions matched, check registered functions
	if fn == nil {
		for _, k := range kinds {
			if k.unit == kind {
				fn = k.combine
			}
		}
		if fn == nil {
			return nil, fmt.Errorf("unsupported Op %s", x.Kind)
		}
	}
	lit, err := fn(op, x, y, combine)
	return lit, err
}

func floatOp(op token.Token, x, y *BasicLit, combine bool) (*BasicLit, error) {
	out := &BasicLit{
		Kind: token.FLOAT,
	}
	l, err := strconv.ParseFloat(x.Value, 64)
	if err != nil {
		return out, err
	}
	r, err := strconv.ParseFloat(y.Value, 64)
	if err != nil {
		return out, err
	}
	var t float64
	switch op {
	case token.ADD:
		t = l + r
	case token.SUB:
		t = l - r
	case token.QUO:
		// Sass division can create floats, so much treat
		// ints as floats then test for fitting inside INT
		t = l / r
	case token.MUL:
		t = l * r
	default:
		panic("unsupported intOp" + op.String())
	}
	out.Value = strconv.FormatFloat(t, 'G', -1, 64)
	if math.Remainder(t, 1) == 0 {
		out.Kind = token.INT
	}
	return out, nil

}

func intOp(op token.Token, x, y *BasicLit, combine bool) (*BasicLit, error) {
	out := &BasicLit{
		Kind: x.Kind,
	}
	l, err := strconv.Atoi(x.Value)
	if err != nil {
		return out, err
	}
	r, err := strconv.Atoi(y.Value)
	if err != nil {
		return out, err
	}
	var t int
	switch op {
	case token.ADD:
		t = l + r
	case token.SUB:
		t = l - r
	case token.QUO:
		// Sass division can create floats, so much treat
		// ints as floats then test for fitting inside INT
		fl, fr := float64(l), float64(r)
		if math.Remainder(fl, fr) != 0 {
			return floatOp(op, x, y, combine)
		}
		t = l / r
	case token.MUL:
		t = l * r
	default:
		panic("unsupported intOp" + op.String())
	}
	out.Value = strconv.Itoa(t)
	return out, nil
}

func stringOp(op token.Token, x, y *BasicLit, combine bool) (*BasicLit, error) {
	kind := token.STRING
	if op == token.ADD {
		return &BasicLit{
			Kind:  kind,
			Value: x.Value + y.Value,
		}, nil
	}

	return &BasicLit{
		Kind:  kind,
		Value: x.Value + op.String() + y.Value,
	}, nil
}
