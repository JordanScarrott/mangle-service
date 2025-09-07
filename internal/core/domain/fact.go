package domain

import "github.com/google/mangle/ast"

// Fact represents a fact in the Mangle sense, which is a grounded atom.
// We are aliasing ast.Atom from the Mangle library.
type Fact = ast.Atom
