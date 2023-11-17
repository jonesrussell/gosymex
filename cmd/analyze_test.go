package cmd

import (
	"go/ast"
	"reflect"
	"testing"
)

func TestParseFile(t *testing.T) {
	// Define a test case
	testCase := struct {
		name     string
		filePath string
		wantErr  bool
	}{
		name:     "Test with a valid file",
		filePath: "test_files/testfile.go",
		wantErr:  false,
	}

	t.Run(testCase.name, func(t *testing.T) {
		_, err := parseFile(testCase.filePath)
		if (err != nil) != testCase.wantErr {
			t.Errorf("parseFile() error = %v, wantErr %v", err, testCase.wantErr)
		}
	})
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
			handleImportSpec(testCase.input, details)

			// Check that the Imports field in the details struct matches what we expect
			if !reflect.DeepEqual(details.Imports, testCase.want) {
				t.Errorf("Imports = %v, want %v", details.Imports, testCase.want)
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
				Imports:  []string{"fmt", "net/http"}, // Update this to match the imports in testfile.go
				Structs: map[string][]string{
					"MyStruct": {"Field1 int", "Field2 string"}, // Update this to match the structs in testfile.go
				},
				Funcs: []string{
					"MyFunc(param1 int, param2 string) returns (result bool)", // Update this to match the functions in testfile.go
					"main()",
				},
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
			}
		})
	}
}
