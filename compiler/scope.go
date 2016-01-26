package compiler

import (
	"fmt"

	"github.com/wellington/sass/ast"
)

// stores types and values with scoping. To remove a scope
// use CloseScope(), to open a new Scope use OpenScope().
type Scope interface {
	// OpenScope() Typ
	// CloseScope() Typ
	Get(string) interface{}
	Set(string, interface{})
	// Number of Rules in this scope
	RuleAdd(*ast.RuleSpec)
	RuleLen() int

	RegisterMixin(string, int, *MixFn)
	Mixin(string, int) (*MixFn, error)
}

var (
	empty = new(emptyTyp)
)

type emptyTyp struct{}

func (*emptyTyp) Get(name string) interface{} {
	return nil
}

func (*emptyTyp) RegisterMixin(_ string, _ int, _ *MixFn) {}

func (*emptyTyp) Mixin(_ string, _ int) (*MixFn, error) {
	return nil, ErrMixinNotFound
}

func (*emptyTyp) Set(name string, _ interface{}) {}

func (*emptyTyp) RuleLen() int { return 0 }

func (*emptyTyp) RuleAdd(*ast.RuleSpec) {}

type valueScope struct {
	Scope
	rules []*ast.RuleSpec
	m     map[string]interface{}
}

func (t *valueScope) RuleAdd(rule *ast.RuleSpec) {
	t.rules = append(t.rules, rule)
}

func (t *valueScope) RuleLen() int {
	return len(t.rules)
}

func (t *valueScope) Get(name string) interface{} {
	val, ok := t.m[name]
	if ok {
		return val
	}
	return t.Scope.Get(name)
}

func (t *valueScope) Set(name string, v interface{} /* should this just be string? */) {
	fmt.Printf("setting %12s: %-10v\n", name, v)
	t.m[name] = v
}

func NewScope(s Scope) Scope {
	return &valueScope{Scope: s, m: make(map[string]interface{})}
}

func CloseScope(typ Scope) Scope {
	s, ok := typ.(*valueScope)
	if !ok {
		return typ
	}
	return s.Scope
}
