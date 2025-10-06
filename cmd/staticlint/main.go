package main

import (
	"github.com/argad/url-shortener/cmd/staticlint/exitcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	// Create a slice to store all analyzers
	var analyzers []*analysis.Analyzer

	// Add standard analyzers
	analyzers = append(analyzers,
		printf.Analyzer,    // Check fmt.Printf
		shadow.Analyzer,    // Check shadowed variables
		structtag.Analyzer, // Check struct tags
	)

	// Add SA class analyzers from staticcheck
	for _, v := range staticcheck.Analyzers {
		if v.Analyzer.Name[:2] == "SA" {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// Add one analyzer from each other staticcheck class
	// For example: S1000 (style), T1000 (simplifications), QF1000 (quickfixes)
	checks := map[string]bool{
		"S1000":  true,
		"T1000":  true,
		"QF1000": true,
	}
	for _, v := range staticcheck.Analyzers {
		if checks[v.Analyzer.Name] {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// Add our custom analyzer
	analyzers = append(analyzers, exitcheck.NoExitAnalyzer)

	multichecker.Main(analyzers...)
}
