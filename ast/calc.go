package ast

import (
	"fmt"
	"math"
	"strconv"

	"github.com/wellington/sass/token"
)

func Op(op token.Token, x, y *BasicLit) (*BasicLit, error) {
	fmt.Printf("kind: %s op: %s x: % #v y: % #v\n", x.Kind, op, x, y)
	kind := x.Kind
	switch {
	case kind == token.COLOR:
		return colorOp(op, x, y)
	case kind == token.INT:
		switch y.Kind {
		case token.STRING:
			return stringOp(op, x, y)
		case token.FLOAT:
			return floatOp(op, x, y)
		}
		return intOp(op, x, y)
	case kind == token.FLOAT:
		switch y.Kind {
		case token.STRING:
			return stringOp(op, x, y)
		}
		return floatOp(op, x, y)
	case kind == token.STRING || x.Kind != y.Kind:
		return stringOp(op, x, y)
	default:
		return nil, fmt.Errorf("unsupported Op %s", x.Kind)
	}
}

func floatOp(op token.Token, x, y *BasicLit) (*BasicLit, error) {
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

func intOp(op token.Token, x, y *BasicLit) (*BasicLit, error) {
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
			return floatOp(op, x, y)
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

func stringOp(op token.Token, x, y *BasicLit) (*BasicLit, error) {
	if op == token.ADD {
		return &BasicLit{
			Kind:  token.STRING,
			Value: x.Value + y.Value,
		}, nil
	}
	return &BasicLit{
		Kind:  token.STRING,
		Value: x.Value + op.String() + y.Value,
	}, nil
}
