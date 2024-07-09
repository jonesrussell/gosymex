package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe a Go file",
	Long:  `This command describes a Go file and prints out its details.`,
	Run:   runDescribeCmd,
}

func init() {
	describeCmd.Flags().BoolP("include-tests", "t", false, "Include test files in the recursive describe")
	describeCmd.Flags().BoolP("include-mocks", "m", false, "Include mock files in the recursive describe")
	rootCmd.AddCommand(describeCmd)
}

func runDescribeCmd(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: gosymex describe <filepath or directory>")
		return
	}

	path := args[0]
	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Println("Error accessing path:", err)
		return
	}

	includeTests, _ := cmd.Flags().GetBool("include-tests")
	includeMocks, _ := cmd.Flags().GetBool("include-mocks")

	if fileInfo.IsDir() {
		processDirectory(path, includeTests, includeMocks)
	} else {
		describeFile(path)
	}
}

func processDirectory(path string, includeTests bool, includeMocks bool) {
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if isGoFile(filePath, info, includeTests, includeMocks) {
			describeFile(filePath)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error walking the directory:", err)
	}
}

func isGoFile(filePath string, info os.FileInfo, includeTests bool, includeMocks bool) bool {
	return !info.IsDir() && strings.HasSuffix(filePath, ".go") && (includeTests || !strings.HasSuffix(filePath, "_test.go")) && (includeMocks || !strings.HasSuffix(filePath, "_mock.go"))
}

func describeFile(filePath string) (string, error) {
	// Check if the file is a Go file
	if filepath.Ext(filePath) != ".go" {
		// If not, return a JSON object with an error field
		return `{"error":"not a Go file"}`, errors.New("not a Go file")
	}

	// Parse the Go file at the given path
	node, err := parseFile(filePath)
	if err != nil {
		// If there's an error, return an empty string and the error
		return "", fmt.Errorf("Error parsing file: %v", err)
	}

	// Inspect the AST of the Go file and get its details
	details := inspectFile(filePath, node)

	// Marshal the details into a JSON string
	jsonDetails, _ := json.MarshalIndent(details, "", "  ")

	// Return the JSON string and nil error
	return string(jsonDetails), nil
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

// inspectFile inspects the AST of a Go file and returns a FileDetails struct.
func inspectFile(filePath string, node *ast.File) *FileDetails {
	details := &FileDetails{
		FilePath:   filePath,
		Imports:    []string{},
		Structs:    make(map[string][]string),
		Interfaces: nil,
		Funcs:      []string{},
	}

	handlers := map[string]func(ast.Node, *FileDetails){
		"*ast.ImportSpec": handleImportSpec,
		"*ast.TypeSpec":   handleTypeSpec,
		"*ast.FuncDecl":   handleFuncDecl,
	}

	ast.Inspect(node, func(n ast.Node) bool {
		handler, ok := handlers[fmt.Sprintf("%T", n)]
		if ok {
			handler(n, details)
		}
		return true
	})

	return details
}

// handleImportSpec handles an import spec AST node.
func handleImportSpec(n ast.Node, details *FileDetails) {
	x, ok := n.(*ast.ImportSpec)
	if !ok {
		return // or handle the error as you see fit
	}
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

func handleTypeSpec(n ast.Node, details *FileDetails) {
	x, ok := n.(*ast.TypeSpec)
	if !ok {
		return // or handle the error as you see fit
	}
	// The rest of your function implementation remains the same
	switch t := x.Type.(type) {
	case *ast.StructType:
		// Add an entry for the struct to the Structs field
		details.Structs[x.Name.Name] = []string{}

		// Then add each field to the entry
		for _, f := range t.Fields.List {
			if len(f.Names) > 0 { // Check if the Names slice is not empty
				field := fmt.Sprintf("%s %s", f.Names[0].Name, types.ExprString(f.Type))
				details.Structs[x.Name.Name] = append(details.Structs[x.Name.Name], field)
			}
		}
	}
}

// handleFuncDecl handles a function declaration AST node.
func handleFuncDecl(n ast.Node, details *FileDetails) {
	x, ok := n.(*ast.FuncDecl)
	if !ok {
		return // or handle the error as you see fit
	}
	// The rest of your function implementation remains the same
	funcSig := ""
	if x.Recv != nil { // Check if the function has a receiver
		// Assuming the receiver is a single field, extract the type
		receiverType := types.ExprString(x.Recv.List[0].Type)
		funcSig += fmt.Sprintf("(%s).", receiverType)
	}
	funcSig += fmt.Sprintf("%s(", x.Name.Name)
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
