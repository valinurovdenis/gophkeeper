package main

import (
	"github.com/eltonjr/json-interface-linter/jsontag"
	"github.com/eltonjr/json-interface-linter/marshal"
	"github.com/ichiban/cyclomatic"
	"github.com/valinurovdenis/gophkeeper/internal/staticlint/exitchecker"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck/st1000"
)

// Runs analyzers check for files in given directory
// ./staticlint ./...
//
// Analyzers:
//
//	all SA analyzers from staticcheck package
//	printf		check print template arguments
//	shadow		check shadowed variables
//	structtag   check struct field tags are well formed
//	atomic		check proper use of atomics
//	st1000		check correct package comment
//	cyclomatic	check functions complexity
//	jsontag		check if structs tagged as json contain an interface
//	marshal		check if marshaled structs contain an interface
//	ExitCheckAnalyzer checks for os.Exit call from main
func main() {
	analyzers := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		atomic.Analyzer,
		st1000.Analyzer,
		cyclomatic.Analyzer,
		jsontag.Analyzer,
		marshal.Analyzer,
		exitchecker.ExitCheckAnalyzer,
	}
	for _, v := range staticcheck.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}
	multichecker.Main(analyzers...)
}
