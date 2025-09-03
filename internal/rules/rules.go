// Package rules defines the Datalog rules for the microservice.
package rules

// DependsOnRules defines the dependency relationship, including transitivity.
const DependsOnRules = `
  depends_on(X, Y) :- calls(X, Y).
  depends_on(X, Z) :- calls(X, Y), depends_on(Y, Z).
`
