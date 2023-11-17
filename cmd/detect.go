package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
)

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
}

type dependency struct {
	Name     string
	Version  string
	Indirect bool
}

// Update detectRun function
func detectRun(cmd *cobra.Command, args []string) {
	path := args[0]

	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Error accessing path '%s': %v\n", path, err)
		return
	}

	if info.Mode().IsRegular() {
		path = filepath.Dir(path)
	}

	for {
		goModPath := filepath.Join(path, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			modulePath, dependencies, err := readGoModFile(goModPath)
			if err != nil {
				fmt.Printf("Error reading go.mod file: %v\n", err)
				return
			}
			printProjectDetails(path, modulePath, dependencies)
			return
		}

		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		path = parent
	}

	fmt.Printf("'%s' is not a Go project. No go.mod file found.\n", args[0])
}

// Update printProjectDetails function
func printProjectDetails(projectPath, modulePath string, dependencies []dependency) {
	// Create a new table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// Set style to StyleColoredBlackOnGreenWhite
	t.SetStyle(table.StyleColoredBlackOnGreenWhite)

	// Print project information in the header
	t.AppendHeader(table.Row{"  Project Name:", filepath.Base(projectPath)})
	t.AppendHeader(table.Row{"  Module Path:", modulePath})
	t.AppendHeader(table.Row{"  Type:", "Go project"})

	// Set style back to StyleColoredGreenWhiteOnBlack for dependencies section
	t.SetStyle(table.StyleColoredGreenWhiteOnBlack)

	// Set column headers for dependencies
	t.AppendSeparator()
	t.AppendHeader(table.Row{"#", "Dependency", "Version", "Indirect"})

	// Populate the table with data
	for i, dep := range dependencies {
		t.AppendRow([]interface{}{i + 1, dep.Name, dep.Version, dep.Indirect})
	}

	// Render the table
	t.Render()
}

func readGoModFile(goModPath string) (string, []dependency, error) {
	file, err := os.Open(goModPath)
	if err != nil {
		return "", nil, fmt.Errorf("error opening go.mod file: %v", err)
	}
	defer file.Close()

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

func parseRequire(requires ...*modfile.Require) []dependency {
	var dependencies []dependency
	for _, require := range requires {
		dependencies = append(dependencies, dependency{
			Name:     require.Mod.Path,
			Version:  require.Mod.Version,
			Indirect: require.Indirect,
		})
	}
	return dependencies
}
