// Cтандартные статические анализаторы пакета golang.org/x/tools/go/analysis/passes
// asmdecl
// Package asmdecl defines an Analyzer that reports mismatches between assembly files and Go declarations.
// assign
// Package assign defines an Analyzer that detects useless assignments.
// atomic
// Package atomic defines an Analyzer that checks for common mistakes using the sync/atomic package.
// atomicalign
// Package atomicalign defines an Analyzer that checks for non-64-bit-aligned arguments to sync/atomic functions.
// bools
// Package bools defines an Analyzer that detects common mistakes involving boolean operators.
// buildtag
// Package buildtag defines an Analyzer that checks build tags.
// composite
// Package composite defines an Analyzer that checks for unkeyed composite literals.
// copylock
// Package copylock defines an Analyzer that checks for locks erroneously passed by value.
// errorsas
// The errorsas package defines an Analyzer that checks that the second argument to errors.As is a pointer to a type implementing error.
// fieldalignment
// Package fieldalignment defines an Analyzer that detects structs that would use less memory if their fields were sorted.
// httpresponse
// Package httpresponse defines an Analyzer that checks for mistakes using HTTP responses.
// ifaceassert
// Package ifaceassert defines an Analyzer that flags impossible interface-interface type assertions.
// inspect
// Package inspect defines an Analyzer that provides an AST inspector (golang.org/x/tools/go/ast/inspector.Inspector) for the syntax trees of a package.
// loopclosure
// Package loopclosure defines an Analyzer that checks for references to enclosing loop variables from within nested functions.
// lostcancel
// Package lostcancel defines an Analyzer that checks for failure to call a context cancellation function.
// nilfunc
// Package nilfunc defines an Analyzer that checks for useless comparisons against nil.
// printf
// Package printf defines an Analyzer that checks consistency of Printf format strings and arguments.
// shadow
// Package shadow defines an Analyzer that checks for shadowed variables.
// shift
// Package shift defines an Analyzer that checks for shifts that exceed the width of an integer.
// sigchanyzer
// Package sigchanyzer defines an Analyzer that detects misuse of unbuffered signal as argument to signal.Notify.
// sortslice
// Package sortslice defines an Analyzer that checks for calls to sort.Slice that do not use a slice type as first argument.
// stdmethods
// Package stdmethods defines an Analyzer that checks for misspellings in the signatures of methods similar to well-known interfaces.
// stringintconv
// Package stringintconv defines an Analyzer that flags type conversions from integers to strings.
// structtag
// Package structtag defines an Analyzer that checks struct field tags are well formed.
// testinggoroutine
// Package testinggoroutine defines an Analyzerfor detecting calls to Fatal from a test goroutine.
// tests
// Package tests defines an Analyzer that checks for common mistaken usages of tests and examples.
// unmarshal
// The unmarshal package defines an Analyzer that checks for passing non-pointer or non-interface types to unmarshal and decode functions.
// unreachable
// Package unreachable defines an Analyzer that checks for unreachable code.
// unsafeptr
// Package unsafeptr defines an Analyzer that checks for invalid conversions of uintptr to unsafe.Pointer.
// unusedresult
// Package unusedresult defines an analyzer that checks for unused results of calls to certain pure functions.
// unusedwrite
// Package unusedwrite checks for unused writes to the elements of a struct or array object.
// usesgenerics
// Package usesgenerics defines an Analyzer that checks for usage of generic features added in Go 1.18.

// Staticcheck
// Staticcheck is a state of the art linter for the Go programming language. Using static analysis, it finds bugs and performance issues, offers simplifications, and enforces style rules.
// Each of the 150+ checks has been designed to be fast, precise and useful. When Staticcheck flags code, you can be sure that it isn’t wasting your time with unactionable warnings. Unlike many other linters, Staticcheck focuses on checks that produce few to no false positives. It’s the ideal candidate for running in CI without risking spurious failures.
// Staticcheck aims to be trivial to adopt. It behaves just like the official go tool and requires no learning to get started with. Just run staticcheck ./... on your code in addition to go vet ./....
// Staticcheck can be used from the command line, in CI, and even directly from your editor.
// Description of ALL CHECKS are avaliable on https://staticcheck.io/docs/checks/

// errcheck
// errcheck is a program for checking for unchecked errors in Go code.
// Use
// For basic usage, just give the package path of interest as the first argument:
// errcheck github.com/kisielk/errcheck/testdata
// To check all packages beneath the current directory:
// errcheck ./...
// Or check all packages in your $GOPATH and $GOROOT:
// errcheck all

// goone
// goone finds N+1(strictly speaking call SQL in a for loop) query in go

package main
