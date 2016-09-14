package ast

import "strings"

// JoinLits accepts a series of lits and optional separator to
// create a string. It's possible this outputs improper output
// for compiler settings
func JoinLits(a []*BasicLit, sep string) string {
	s := make([]string, len(a))
	for i := range a {
		s[i] = a[i].Value
	}
	return strings.Join(s, sep)
}
