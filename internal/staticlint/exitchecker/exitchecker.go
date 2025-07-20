// Package exitchecker contains analyzer check for os.Exit in main function.
package exitchecker

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Exit function analyzer
var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for os.Exit in main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			for _, d := range file.Decls {
				fn, isFunc := d.(*ast.FuncDecl)
				if isFunc && fn.Name.Name == "main" {
					ast.Inspect(fn, func(node ast.Node) bool {
						call, isCall := node.(*ast.CallExpr)
						if !isCall {
							return true
						}

						sel, isSelector := call.Fun.(*ast.SelectorExpr)
						if !isSelector {
							return true
						}

						pkg, isPkg := sel.X.(*ast.Ident)
						if isPkg && pkg.Name == "os" && sel.Sel.Name == "Exit" {
							pass.Reportf(node.Pos(), "os.Exit call in main function")
						}
						return true
					})
				}
			}
		}
	}

	return nil, nil
}
