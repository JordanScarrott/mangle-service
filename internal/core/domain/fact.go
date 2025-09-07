package domain

import "github.com/google/mangle/ast"

// Fact represents a fact in the Mangle sense, which is a grounded atom.
// We are aliasing ast.Atom from the Mangle library.
type Fact = ast.Atom

// FactsToClauses converts a slice of Fact (ast.Atom) to a slice of ast.Clause.
func FactsToClauses(facts []Fact) []ast.Clause {
	clauses := make([]ast.Clause, len(facts))
	for i, fact := range facts {
		clauses[i] = ast.Clause{Head: fact}
	}
	return clauses
}
