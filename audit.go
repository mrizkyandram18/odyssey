package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fset := token.NewFileSet()
	out, err := os.Create("results_utf8.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer out.Close()

	var count int
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.Contains(path, "audit.go") || strings.Contains(path, "node_modules") {
			return nil
		}
		
		node, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil
		}

		ast.Inspect(node, func(n ast.Node) bool {
			// Look for if err != nil
			ifStmt, ok := n.(*ast.IfStmt)
			if ok {
				isErrCheck := false
				
				// Very basic heuristic: check if condition contains "err"
				ast.Inspect(ifStmt.Cond, func(cn ast.Node) bool {
					if ident, iok := cn.(*ast.Ident); iok && ident.Name == "err" {
						isErrCheck = true
					}
					return true
				})

				if isErrCheck && ifStmt.Body != nil {
					// Check for continue or return nil inside the block
					for _, stmt := range ifStmt.Body.List {
						switch st := stmt.(type) {
						case *ast.BranchStmt:
							if st.Tok == token.CONTINUE {
								fmt.Fprintf(out, "File: %s:%d\nFound: continue in err check\n\n", path, fset.Position(ifStmt.Pos()).Line)
								count++
							}
						case *ast.ReturnStmt:
							if len(st.Results) == 1 {
								if ident, iok := st.Results[0].(*ast.Ident); iok && ident.Name == "nil" {
									fmt.Fprintf(out, "File: %s:%d\nFound: return nil in err check\n\n", path, fset.Position(ifStmt.Pos()).Line)
									count++
								}
							}
							if len(st.Results) == 0 {
								fmt.Fprintf(out, "File: %s:%d\nFound: empty return in err check\n\n", path, fset.Position(ifStmt.Pos()).Line)
								count++
							}
						}
					}
				}
			}

            // Look for recover/panic
            callExpr, ok := n.(*ast.CallExpr)
            if ok {
                if ident, iok := callExpr.Fun.(*ast.Ident); iok {
                    if ident.Name == "recover" || ident.Name == "panic" {
                        fmt.Fprintf(out, "File: %s:%d\nFound: %s\n\n", path, fset.Position(callExpr.Pos()).Line, ident.Name)
                        count++
                    }
                }
            }
            
            // Look for type assertions that are unsafe (without ok)
            typeAssert, ok := n.(*ast.TypeAssertExpr)
            if ok {
                // Check if it's part of a safe assignment (v, ok := ...)
                // This is a bit complex in AST, we will just flag type assertions and review manually if count is small.
                // To keep noise low, let's skip for now unless we need to.
                _ = typeAssert
            }
			return true
		})
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(out, "Total: %d\n", count)
}
