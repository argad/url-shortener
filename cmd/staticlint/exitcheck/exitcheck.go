package exitcheck

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// NoExitAnalyzer create the Analyzer instance
var NoExitAnalyzer = &analysis.Analyzer{
	Name: "noexit",
	Doc:  "check for direct os.Exit calls in main function of main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			// Find the main function
			funcDecl, ok := node.(*ast.FuncDecl)
			if !ok || funcDecl.Name.Name != "main" {
				return true
			}

			// Check for os.Exit calls inside the main function
			ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
				callExpr, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				// Use TypesInfo.Uses to get the actual object being referenced
				if obj := pass.TypesInfo.Uses[selectorExpr.Sel]; obj != nil {
					// Check if this is a function from the "os" package
					if pkg := obj.Pkg(); pkg != nil && pkg.Path() == "os" && obj.Name() == "Exit" {
						pass.Reportf(callExpr.Pos(), "direct call to os.Exit in main function is forbidden")
					}
				}

				return true
			})

			return false // Found main, no need to inspect further
		})
	}

	return nil, nil
}
