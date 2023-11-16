/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strings"

	"github.com/spf13/cobra"
)

type FileDetails struct {
	FilePath string
	Imports  []string
	Structs  map[string][]string
	Funcs    []string
}

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze a Go file",
	Long:  `This command analyzes a Go file and prints out its details.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Usage: gosymex analyze <filepath>")
			return
		}

		filePath := args[0]

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			fmt.Println(err)
			return
		}

		details := FileDetails{
			FilePath: filePath,
			Structs:  make(map[string][]string),
		}

		ast.Inspect(node, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.ImportSpec:
				// Remove quotes from import paths
				importPath := strings.Trim(x.Path.Value, "\"")
				details.Imports = append(details.Imports, importPath)
			case *ast.TypeSpec:
				switch t := x.Type.(type) {
				case *ast.StructType:
					for _, f := range t.Fields.List {
						field := fmt.Sprintf("%s %s", f.Names[0].Name, types.ExprString(f.Type))
						details.Structs[x.Name.Name] = append(details.Structs[x.Name.Name], field)
					}
				}
			case *ast.FuncDecl:
				funcSig := fmt.Sprintf("%s(", x.Name.Name)
				if x.Type.Params != nil {
					for i, p := range x.Type.Params.List {
						if i > 0 {
							funcSig += ", "
						}
						for j := range p.Names {
							if j > 0 {
								funcSig += ", "
							}
							funcSig += fmt.Sprintf("%s %s", p.Names[j], types.ExprString(p.Type))
						}
					}
				}
				funcSig += ")"
				if x.Type.Results != nil {
					funcSig += " returns ("
					for i, r := range x.Type.Results.List {
						if i > 0 {
							funcSig += ", "
						}
						if len(r.Names) > 0 {
							funcSig += fmt.Sprintf("%s ", r.Names[0])
						}
						funcSig += types.ExprString(r.Type)
					}
					funcSig += ")"
				}
				details.Funcs = append(details.Funcs, funcSig)
			}
			return true
		})

		jsonDetails, _ := json.MarshalIndent(details, "", "  ")
		fmt.Println(string(jsonDetails))
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// analyzeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// analyzeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
