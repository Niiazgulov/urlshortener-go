// multichecker - анализатор, состоящий из:
// 1) стандартных статических анализаторов пакета golang.org/x/tools/go/analysis/passes;
// 2) всех анализаторов класса SA пакета staticcheck.io;
// 3) не менее одного анализатора остальных классов пакета staticcheck.io;
// 4) двух публичных анализаторов
// 5) анализатор, запрещающий использовать прямой вызов os.Exit в функции main пакета main.
package main

import (
	"go/ast"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/masibw/goone"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// анализатор os.Exit
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "exitchecker",
	Doc:  "checking of call os Exit in main.go",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			mainDecl, boolDecl := node.(*ast.FuncDecl)
			if !boolDecl {
				return true
			}
			if mainDecl.Name.Name != "main" {
				return false
			}
			ast.Inspect(mainDecl, func(node ast.Node) bool {
				callExpr, boolCallExpr := node.(*ast.CallExpr)
				if !boolCallExpr {
					return true
				}
				s, boolSelectorExpr := callExpr.Fun.(*ast.SelectorExpr)
				if !boolSelectorExpr {
					return true
				}
				if s.Sel.Name == "Exit" {
					ident := s.X.(*ast.Ident)
					if ident.Name == "os" {
						pass.Reportf(s.Pos(), "os exit call")
					}
				}
				return false
			})
			return false
		})
	}
	return nil, nil
}

func main() {
	mychecks := []*analysis.Analyzer{
		asmdecl.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		httpresponse.Analyzer,
		inspect.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		tests.Analyzer,
		testinggoroutine.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		usesgenerics.Analyzer,
		goone.Analyzer,    // внешний анализатор
		errcheck.Analyzer, // внешний анализатор
		OsExitAnalyzer,    // анализатор os.Exit
	}

	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}
	for _, v := range quickfix.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}
	for _, v := range stylecheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}

	multichecker.Main(
		mychecks...,
	)
}
