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

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe a Go file",
	Long:  `This command describes a Go file and prints out its details.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Usage: gosymex describe <filepath>")
			return
		}

		filePath := args[0]
		node, err := parseFile(filePath)
		if err != nil {
			fmt.Println(err)
			return
		}

		details := inspectFile(filePath, node)
		jsonDetails, _ := json.MarshalIndent(details, "", "  ")
		fmt.Println(string(jsonDetails))
	},
}

func init() {
	rootCmd.AddCommand(describeCmd)
}

type FileDetails struct {
	FilePath   string
	Imports    []string
	Structs    map[string][]string
	Interfaces map[string][]string
	Funcs      []string
}

// parseFile parses the Go file at the given path and returns the corresponding AST node.
func parseFile(filePath string) (*ast.File, error) {
	fset := token.NewFileSet()
	return parser.ParseFile(fset, filePath, nil, parser.ParseComments)
}

// handleImportSpec handles an import spec AST node.
func handleImportSpec(x *ast.ImportSpec, details *FileDetails) {
	importPath := strings.Trim(x.Path.Value, "\"")
	details.Imports = append(details.Imports, importPath)
}

// handleInterfaceSpec handles an interface spec AST node.
func handleInterfaceSpec(x *ast.TypeSpec, details *FileDetails) {
	switch t := x.Type.(type) {
	case *ast.InterfaceType:
		// Add an entry for the interface to the Interfaces field
		details.Interfaces[x.Name.Name] = []string{}

		// Then add each method to the entry
		for _, f := range t.Methods.List {
			method := fmt.Sprintf("%s %s", f.Names[0].Name, types.ExprString(f.Type))
			details.Interfaces[x.Name.Name] = append(details.Interfaces[x.Name.Name], method)
		}
	}
}

// handleTypeSpec handles a type spec AST node.
func handleTypeSpec(x *ast.TypeSpec, details *FileDetails) {
	switch t := x.Type.(type) {
	case *ast.StructType:
		// Add an entry for the struct to the Structs field
		details.Structs[x.Name.Name] = []string{}

		// Then add each field to the entry
		for _, f := range t.Fields.List {
			field := fmt.Sprintf("%s %s", f.Names[0].Name, types.ExprString(f.Type))
			details.Structs[x.Name.Name] = append(details.Structs[x.Name.Name], field)
		}
	}
}

// handleFuncDecl handles a function declaration AST node.
func handleFuncDecl(x *ast.FuncDecl, details *FileDetails) {
	funcSig := ""
	if x.Recv != nil { // Check if the function has a receiver
			// Assuming the receiver is a single field, extract the type
			receiverType := types.ExprString(x.Recv.List[0].Type)
			funcSig += fmt.Sprintf("(%s).", receiverType)
	}
	funcSig += fmt.Sprintf("%s(", x.Name.Name)
	if x.Type.Params != nil {
			for i, p := range x.Type.Params.List {
					if i >  0 {
							funcSig += ", "
					}
					for j := range p.Names {
							if j >  0 {
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
					if i >  0 {
							funcSig += ", "
					}
					if len(r.Names) >  0 {
							funcSig += fmt.Sprintf("%s ", r.Names[0])
					}
					funcSig += types.ExprString(r.Type)
			}
			funcSig += ")"
	}
	details.Funcs = append(details.Funcs, funcSig)
}

// inspectFile inspects the AST of a Go file and returns a FileDetails struct.
func inspectFile(filePath string, node *ast.File) *FileDetails {
	details := &FileDetails{
		FilePath: filePath,
		Imports:  []string{},
		Interfaces: make(map[string][]string),
		Structs:  make(map[string][]string),
		Funcs:    []string{},
	}

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.ImportSpec:
			handleImportSpec(x, details)
		case *ast.TypeSpec:
			if _, ok := x.Type.(*ast.InterfaceType); ok {
				handleInterfaceSpec(x, details)
			} else {
				handleTypeSpec(x, details)
			}
		case *ast.FuncDecl:
			handleFuncDecl(x, details)
		}
		return true
	})

	return details
}
