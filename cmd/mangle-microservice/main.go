package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"mangle-microservice-example/internal/rules"

	"github.com/google/mangle/analysis"
	"github.com/google/mangle/ast"
	"github.com/google/mangle/engine"
	"github.com/google/mangle/factstore"
	"github.com/google/mangle/parse"
)

func main() {
	// A fact is a clause with an empty body.
	fact1, err := parse.Clause(`calls("service-a", "service-b").`)
	if err != nil {
		log.Fatalf("failed to parse fact1: %v", err)
	}
	fact2, err := parse.Clause(`calls("service-b", "service-c").`)
	if err != nil {
		log.Fatalf("failed to parse fact2: %v", err)
	}

	// Rules are also parsed from strings.
	// We import the rules from our internal package.
	rulesUnit, err := parse.Unit(strings.NewReader(rules.DependsOnRules))
	if err != nil {
		log.Fatalf("failed to parse rules: %v", err)
	}

	// A SourceUnit contains declarations and clauses.
	// We include both our facts and our rules.
	sourceUnit := parse.SourceUnit{
		Clauses: append([]ast.Clause{fact1, fact2}, rulesUnit.Clauses...),
	}

	// The analyzer takes a source unit and produces a program.
	// The program can then be evaluated.
	program, err := analysis.AnalyzeOneUnit(sourceUnit, nil)
	if err != nil {
		log.Fatalf("failed to create program: %v", err)
	}

	// We create an empty in-memory fact store.
	store := factstore.NewSimpleInMemoryStore()

	// We evaluate the program. This populates the fact store.
	err = engine.EvalProgram(program, store)
	if err != nil {
		log.Fatalf("program evaluation failed: %v", err)
	}

	// The query is an atom.
	// We are looking for all services that depend on "service-c".
	query, err := parse.Atom(`depends_on(X, "service-c")`)
	if err != nil {
		log.Fatalf("failed to parse query: %v", err)
	}

	// We can now query the fact store for the results.
	// The results are returned as a list of facts.
	fmt.Printf("Query: %s?\n", query.String())
	var results []ast.Atom
	store.GetFacts(query, func(a ast.Atom) error {
		results = append(results, a)
		return nil
	})

	fmt.Println("Results:")
	// We iterate over the results and print them.
	var resultVars []string
	for _, fact := range results {
		// The arguments of the fact correspond to the variables in the query.
		// In our query `depends_on(X, "service-c")?`, X is the first argument.
		resultVars = append(resultVars, fact.Args[0].String())
	}
	// The order of results is not guaranteed, so we sort them for stable output.
	// This is not strictly necessary but makes testing easier.
	sort.Strings(resultVars)
	for _, val := range resultVars {
		fmt.Printf("  X = %s\n", val)
	}
}
