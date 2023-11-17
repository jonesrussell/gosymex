package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
)

// detectCmd represents the detect command
var detectCmd = &cobra.Command{
	Use:   "detect [path]",
	Short: "Detect if a file or directory is part of a Go project and report basic info",
	Args:  cobra.ExactArgs(1),
	Run:   detectRun,
}

var showAllDeps bool

func init() {
	rootCmd.AddCommand(detectCmd)

	detectCmd.Flags().BoolVar(&showAllDeps, "all-deps", false, "Show all dependencies")

	// You can define flags for the detect command here if needed.
	// For example:
	// detectCmd.Flags().StringP("output", "o", "", "Specify output format")
}

func detectRun(cmd *cobra.Command, args []string) {
	path := args[0]

	// If the given path is a file, get its directory
	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Error accessing path '%s': %v\n", path, err)
		return
	}

	if info.Mode().IsRegular() {
		path = filepath.Dir(path)
	}

	// Recursively check parent directories for go.mod file
	for {
		goModPath := filepath.Join(path, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			fmt.Printf("'%s' is a Go project.\n", path)
			modulePath, dependencies, err := readGoModFile(goModPath)
			if err != nil {
				fmt.Printf("Error reading go.mod file: %v\n", err)
				return
			}
			printProjectDetails(path, modulePath, dependencies)
			return
		}

		// Move up one directory
		parent := filepath.Dir(path)
		if parent == path {
			// Reached the root directory without finding go.mod
			break
		}
		path = parent
	}

	fmt.Printf("'%s' is not a Go project. No go.mod file found.\n", args[0])
}

// struct to store dependency details
type dependency struct {
	Name    string
	Version string
	Direct  bool
}

func printProjectDetails(projectPath, modulePath string, dependencies []dependency) {
	fmt.Println("Project Details:")
	// Extract project name from the directory name
	projectName := filepath.Base(projectPath)
	fmt.Printf("  Project Name: %s\n", projectName)

	fmt.Printf("  Module Path: %s\n", modulePath)

	if len(dependencies) > 0 {
		fmt.Println("  Dependencies:")
		for _, dep := range dependencies {
			if showAllDeps || dep.Direct {
				fmt.Printf("    %s %s\n", dep.Name, dep.Version)
			}
		}
	} else {
		fmt.Println("  No dependencies found.")
	}
	// Additional details can be added based on your requirements
}

// readGoModFile reads the go.mod file and returns module path and dependencies
func readGoModFile(goModPath string) (string, []dependency, error) {
	// Open the go.mod file
	file, err := os.Open(goModPath)
	if err != nil {
		return "", nil, fmt.Errorf("error opening go.mod file: %v", err)
	}
	defer file.Close()

	// Parse the content of the go.mod file
	content, err := io.ReadAll(file)
	if err != nil {
		return "", nil, fmt.Errorf("error reading go.mod file: %v", err)
	}
	modFile, err := modfile.Parse("go.mod", content, nil)
	if err != nil {
		return "", nil, fmt.Errorf("error parsing go.mod file: %v", err)
	}

	modulePath := modFile.Module.Mod.Path

	var dependencies []dependency
	dependencies = append(dependencies, parseRequire(modFile.Require...)...)

	return modulePath, dependencies, nil
}

// parseRequire takes a slice of modfile.Require and converts it into a slice of dependency
func parseRequire(requires ...*modfile.Require) []dependency {
	var dependencies []dependency
	for _, require := range requires {
		dependencies = append(dependencies, dependency{
			Name:    require.Mod.Path,
			Version: require.Mod.Version,
			Direct:  true,
		})
	}
	return dependencies
}
