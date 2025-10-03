/*
Package staticlint provides a set of tools for static analysis of Go code.

Usage:

    go run cmd/staticlint/main.go [packages]

Included analyzers:

Standard analyzers from golang.org/x/tools/go/analysis/passes:
    - printf: checks format strings correctness in Printf-like functions
    - shadow: detects shadowed variables
    - structtag: validates struct tags

SA class analyzers from staticcheck.io:
    - SA1000: invalid regular expression checks
    - SA1001: invalid template pattern checks
    [and other SA analyzers...]

Additional analyzers from staticcheck.io:
    - S1000: simplification of conditional constructs
    - T1000: simplification of type constructs
    - QF1000: quick fixes

Custom analyzers:
    - exitcheck: prohibits direct usage of os.Exit in the main function of main package

To run the analysis, execute:
    go run cmd/staticlint/main.go ./...
*/

package shortener
