package cmd

import (
	"bytes"
	"go/ast"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

func Test_describeFile(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantOut string
	}{
		{
			name: "Valid file path",
			args: args{
				filePath: "./test_files/testfile.go",
			},
			wantErr: false,
			wantOut: `{
				"FilePath": "./test_files/testfile.go",
				"Imports": [
					"fmt",
					"net/http"
				],
				"Structs": {
					"MyStruct": [
						"Field1 int",
						"Field2 string"
					]
				},
				"Interfaces": null,
				"Funcs": [
					"MyFunc(param1 int, param2 string) returns (result bool)",
					"mainTest()"
				]
			}`,
		},
		{
			name: "Invalid file path",
			args: args{
				filePath: "./test_files/testfile.txt",
			},
			wantErr: true,
		},
		// Add more test cases as needed.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to hold the output
			var buf bytes.Buffer

			// Save the original stdout
			old := os.Stdout

			// Create a temporary file and set it as stdout
			temp, _ := os.CreateTemp("", "")
			os.Stdout = temp

			// Call the function
			describeFile(tt.args.filePath)

			// Copy the contents of the temporary file into our buffer
			temp.Seek(0, 0) // Go to the start of the file
			io.Copy(&buf, temp)

			// Restore the original stdout
			os.Stdout = old

			// Check the output
			got := strings.TrimSpace(buf.String())
			wantOut := strings.TrimSpace(tt.wantOut)
			if (got != wantOut && !tt.wantErr) || (got == "" && tt.wantErr) {
				t.Errorf("describeFile() output = %v, wantErr %v", got, tt.wantErr)
			}

		})
	}
}

func TestParseFile(t *testing.T) {
	// Define a table of test cases
	testCases := []struct {
		name     string
		filePath string // The input to inspectFile
		wantErr  bool
	}{
		{
			name:     "Test with a valid file",
			filePath: "test_files/testfile.go",
			wantErr:  false,
		},
		{
			name:     "Test with an invalid file path",
			filePath: "test_files/non_existent_file.go",
			wantErr:  true,
		},
		{
			name:     "Test with a directory instead of a file",
			filePath: "test_files/",
			wantErr:  true,
		},
		{
			name:     "Test with a file that isn't a Go file",
			filePath: "test_files/testfile.txt",
			wantErr:  true,
		},
	}

	// Run each test case
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := parseFile(testCase.filePath)
			if (err != nil) != testCase.wantErr {
				t.Errorf("parseFile() error = %v, wantErr %v", err, testCase.wantErr)
			}
		})
	}
}

func TestHandleImportSpec(t *testing.T) {
	// Define a table of test cases
	testCases := []struct {
		name  string
		input *ast.ImportSpec // The input to handleImportSpec
		want  []string        // The expected output from handleImportSpec
	}{
		{
			name: "Test with a valid import",
			input: &ast.ImportSpec{
				Path: &ast.BasicLit{
					Value: "\"fmt\"",
				},
			},
			want: []string{"fmt"},
		},
		{
			name: "Test with another valid import",
			input: &ast.ImportSpec{
				Path: &ast.BasicLit{
					Value: "\"net/http\"",
				},
			},
			want: []string{"net/http"},
		},
		{
			name: "Test with a third valid import",
			input: &ast.ImportSpec{
				Path: &ast.BasicLit{
					Value: "\"os\"",
				},
			},
			want: []string{"os"},
		},
		{
			name: "Test with an invalid import",
			input: &ast.ImportSpec{
				Path: &ast.BasicLit{
					Value: "\"nonexistentpackage\"",
				},
			},
			want: []string{"nonexistentpackage"},
		},
	}

	// Run each test case
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create a new FileDetails struct for each test
			details := &FileDetails{
				FilePath: "test_files/testfile.go",
				Imports:  []string{}, // Initialize as an empty slice
				Structs:  make(map[string][]string),
				Funcs:    []string{}, // Initialize as an empty slice
			}

			// Call the function with the test case input
			handleImportSpec(testCase.input, details)

			// Check that the Imports field in the details struct matches what we expect
			if !reflect.DeepEqual(details.Imports, testCase.want) {
				t.Errorf("Imports = %v, want %v", details.Imports, testCase.want)
			}
		})
	}
}

func TestHandleInterfaceSpec(t *testing.T) {
	// Define a table of test cases
	testCases := []struct {
		name  string
		input *ast.TypeSpec       // The input to handleInterfaceSpec
		want  map[string][]string // The expected output from handleInterfaceSpec
	}{
		{
			name: "Test with a valid interface type",
			input: &ast.TypeSpec{
				Name: ast.NewIdent("MyInterface"),
				Type: &ast.InterfaceType{
					Methods: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{ast.NewIdent("Method1")},
								Type:  ast.NewIdent("int"),
							},
							{
								Names: []*ast.Ident{ast.NewIdent("Method2")},
								Type:  ast.NewIdent("string"),
							},
						},
					},
				},
			},
			want: map[string][]string{
				"MyInterface": {"Method1 int", "Method2 string"},
			},
		},
		// Add more test cases as needed...
	}

	// Run each test case
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create a new FileDetails struct for each test
			details := &FileDetails{
				FilePath:   "test_files/testfile.go",
				Interfaces: make(map[string][]string),
			}

			// Call the function with the test case input
			handleInterfaceSpec(testCase.input, details)

			// Check that the Interfaces field in the details struct matches what we expect
			if !reflect.DeepEqual(details.Interfaces, testCase.want) {
				t.Errorf("Interfaces = %v, want %v", details.Interfaces, testCase.want)
			}
		})
	}
}

func TestHandleTypeSpec(t *testing.T) {
	// Define a table of test cases
	testCases := []struct {
		name  string
		input *ast.TypeSpec       // The input to handleTypeSpec
		want  map[string][]string // The expected output from handleTypeSpec
	}{
		{
			name: "Test with a valid struct type",
			input: &ast.TypeSpec{
				Name: ast.NewIdent("MyStruct"),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{ast.NewIdent("Field1")},
								Type:  ast.NewIdent("int"),
							},
							{
								Names: []*ast.Ident{ast.NewIdent("Field2")},
								Type:  ast.NewIdent("string"),
							},
						},
					},
				},
			},
			want: map[string][]string{
				"MyStruct": {"Field1 int", "Field2 string"},
			},
		},
		{
			name: "Test with a struct type with no fields",
			input: &ast.TypeSpec{
				Name: ast.NewIdent("EmptyStruct"),
				Type: &ast.StructType{
					Fields: &ast.FieldList{},
				},
			},
			want: map[string][]string{
				"EmptyStruct": {},
			},
		},
		{
			name: "Test with a non-struct type",
			input: &ast.TypeSpec{
				Name: ast.NewIdent("MyInt"),
				Type: ast.NewIdent("int"),
			},
			want: map[string][]string{},
		},
	}

	// Run each test case
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create a new FileDetails struct for each test
			details := &FileDetails{
				FilePath: "test_files/testfile.go",
				Structs:  make(map[string][]string),
			}

			// Call the function with the test case input
			handleTypeSpec(testCase.input, details)

			// Check that the Structs field in the details struct matches what we expect
			if !reflect.DeepEqual(details.Structs, testCase.want) {
				t.Errorf("Structs = %v, want %v", details.Structs, testCase.want)
			}
		})
	}
}

func TestHandleFuncDecl(t *testing.T) {
	// Define a table of test cases
	testCases := []struct {
		name  string
		input *ast.FuncDecl // The input to handleFuncDecl
		want  []string      // The expected output from handleFuncDecl
	}{
		{
			name: "Test with a valid function declaration",
			input: &ast.FuncDecl{
				Name: ast.NewIdent("MyFunc"),
				Type: &ast.FuncType{
					Params: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{ast.NewIdent("param1")},
								Type:  ast.NewIdent("int"),
							},
							{
								Names: []*ast.Ident{ast.NewIdent("param2")},
								Type:  ast.NewIdent("string"),
							},
						},
					},
					Results: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{ast.NewIdent("result")},
								Type:  ast.NewIdent("bool"),
							},
						},
					},
				},
			},
			want: []string{"MyFunc(param1 int, param2 string) returns (result bool)"},
		},
		{
			name: "Test with a function declaration with no parameters",
			input: &ast.FuncDecl{
				Name: ast.NewIdent("NoParamFunc"),
				Type: &ast.FuncType{
					Params: &ast.FieldList{},
					Results: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{ast.NewIdent("result")},
								Type:  ast.NewIdent("bool"),
							},
						},
					},
				},
			},
			want: []string{"NoParamFunc() returns (result bool)"},
		},
		{
			name: "Test with a function declaration with no return values",
			input: &ast.FuncDecl{
				Name: ast.NewIdent("NoReturnFunc"),
				Type: &ast.FuncType{
					Params: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{ast.NewIdent("param1")},
								Type:  ast.NewIdent("int"),
							},
							{
								Names: []*ast.Ident{ast.NewIdent("param2")},
								Type:  ast.NewIdent("string"),
							},
						},
					},
				},
			},
			want: []string{"NoReturnFunc(param1 int, param2 string)"},
		},
	}

	// Run each test case
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create a new FileDetails struct for each test
			details := &FileDetails{
				FilePath: "test_files/testfile.go",
				Structs:  make(map[string][]string),
			}

			// Call the function with the test case input
			handleFuncDecl(testCase.input, details)

			// Check that the Funcs field in the details struct matches what we expect
			if !reflect.DeepEqual(details.Funcs, testCase.want) {
				t.Errorf("Funcs = %v, want %v", details.Funcs, testCase.want)
			}
		})
	}
}

func TestInspectFile(t *testing.T) {
	// Define a table of test cases
	testCases := []struct {
		name     string
		filePath string       // The input to inspectFile
		want     *FileDetails // The expected output from inspectFile
	}{
		{
			name:     "Test with a valid Go file",
			filePath: "test_files/testfile.go",
			want: &FileDetails{
				FilePath: "test_files/testfile.go",
				Imports:  []string{"fmt", "net/http"},
				Structs: map[string][]string{
					"MyStruct": {"Field1 int", "Field2 string"},
				},
				Funcs: []string{
					"MyFunc(param1 int, param2 string) returns (result bool)",
					"mainTest()", // Update this to match the function in testfile.go
				},
			},
		},
		{
			name:     "Test with a Go file that has no imports",
			filePath: "test_files/no_imports.go",
			want: &FileDetails{
				FilePath: "test_files/no_imports.go",
				Imports:  []string{}, // This file has no imports
				Structs:  map[string][]string{},
				Funcs:    []string{"mainNoImports()"}, // This file has a mainNoImports function
			},
		},
		{
			name:     "Test with a Go file that has no structs",
			filePath: "test_files/no_structs.go",
			want: &FileDetails{
				FilePath: "test_files/no_structs.go",
				Imports:  []string{"fmt"},
				Structs:  map[string][]string{},
				Funcs:    []string{"mainNoStructs()"}, // Update this to match the function in no_structs.go
			},
		},
		{
			name:     "Test with a Go file that has no functions",
			filePath: "test_files/no_funcs.go",
			want: &FileDetails{
				FilePath: "test_files/no_funcs.go",
				Imports:  []string{}, // This file has no imports
				Structs: map[string][]string{
					"MyStructNoFuncs": {"Field1 int", "Field2 string"}, // This file declares MyStructNoFuncs
				},
				Funcs: []string{}, // This file has no functions
			},
		},
	}

	// Run each test case
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Parse the Go file to get an *ast.File
			node, err := parseFile(testCase.filePath)
			if err != nil {
				t.Fatalf("parseFile() error = %v", err)
			}

			// Call the function with the test case input
			got := inspectFile(testCase.filePath, node)

			// Check that the returned FileDetails struct matches what we expect
			if !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("inspectFile() = %v, want %v", got, testCase.want)
				t.Logf("got = %#v, want = %#v", got, testCase.want) // Additional logging
			}
		})
	}
}
