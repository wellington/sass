package compiler

import (
	"errors"
	"fmt"

	"github.com/wellington/sass/ast"
)

var (
	ErrMixinNotFound = errors.New("mixin by name not found")
)

type MixFn struct {
	// Minimum arguments that are required to call this mixin
	minArgs int
	maxArgs int
	// Context copied at the creation of mixin, not sure if this
	// is required
	ctx *Context
	fn  *ast.FuncDecl
}

var mixins = map[string]map[int]*MixFn{}

func RegisterMixin(name string, numargs int, fn *MixFn) {
	if mix, ok := mixins[name]; ok {
		if _, ok := mix[numargs]; ok {
			panic(fmt.Sprintf("already registered mixin: %s(%d)",
				name, numargs))
		}
	}
}

func Mixin(name string, numargs int) (*MixFn, error) {
	mixs, ok := mixins[name]
	if !ok {
		return nil, ErrMixinNotFound
	}

	mix, ok := mixs[numargs]
	if !ok {
		return nil, fmt.Errorf("mixin %s with num args %d not found",
			name, numargs)
	}

	return mix, nil
}
