package ast

import "fmt"

// Sort statements for most efficient usage of CSS rules
// (rules first, then other tings)
func StatementsSort(list []Stmt) []Stmt {
	rules := make([]Stmt, 0, len(list))
	notrules := make([]Stmt, 0, len(list))
	for _, stmt := range list {
		switch v := stmt.(type) {
		case *DeclStmt:
			// Rule
			rules = append(rules, stmt)
			fmt.Printf("% #v\n", v.Decl.(*GenDecl).Specs[0].(*RuleSpec).Name)
		case *IncludeStmt, *CommStmt:
			rules = append(rules, stmt)
		case *SelStmt:
			fmt.Printf("% #v\n", v.Name)
			notrules = append(notrules, stmt)
		default:
			notrules = append(notrules, stmt)
		}
	}
	return append(rules, notrules...)
}
