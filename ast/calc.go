package ast

import (
	"fmt"
	"strconv"

	"github.com/wellington/sass/token"
)

func Op(op token.Token, x, y *BasicLit) (*BasicLit, error) {

	if x.Kind != y.Kind || x.Kind == token.STRING {
		return stringOp(op, x, y)
	}

	switch x.Kind {
	case token.INT:
		return intOp(op, x, y)
	case token.COLOR:
		return colorOp(op, x, y)
	default:
		return nil, fmt.Errorf("unsupported Op %s", x.Kind)
	}
}

func intOp(op token.Token, x, y *BasicLit) (*BasicLit, error) {
	out := &BasicLit{
		Kind: op,
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
	return &BasicLit{
		Value: x.Value + op.String() + y.Value,
	}, nil
}
