package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

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
			fmt.Printf("'%s' is a Go project.\n", path)
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

type dependency struct {
	Name     string
	Version  string
	Indirect bool
}

func printProjectDetails(projectPath, modulePath string, dependencies []dependency) {
	fmt.Println("Project Details:")
	projectName := filepath.Base(projectPath)
	fmt.Printf("  Project Name: %s\n", projectName)
	fmt.Printf("  Module Path: %s\n", modulePath)

	// Separate dependencies into direct and indirect
	var directDeps, indirectDeps []dependency
	for _, dep := range dependencies {
		if dep.Indirect {
			indirectDeps = append(indirectDeps, dep)
		} else {
			directDeps = append(directDeps, dep)
		}
	}

	// Sort dependencies (direct first, then indirect)
	sortDependencies(directDeps)
	sortDependencies(indirectDeps)

	if len(dependencies) > 0 {
		fmt.Println("  Dependencies:")
		// Print direct dependencies first
		printDependencies(directDeps, false)
		// Print indirect dependencies next
		printDependencies(indirectDeps, true)
	} else {
		fmt.Println("  No dependencies found.")
	}
}

func sortDependencies(dependencies []dependency) {
	sort.Slice(dependencies, func(i, j int) bool {
		return dependencies[i].Name < dependencies[j].Name
	})
}

func printDependencies(dependencies []dependency, indirect bool) {
	for _, dep := range dependencies {
		if showAllDeps || (!dep.Indirect && !showAllDeps && !indirect) {
			fmt.Printf("    %s, Version: %s", dep.Name, dep.Version)
			if dep.Indirect {
				fmt.Print(" // indirect")
			}
			fmt.Println()
		}
	}
}

func readGoModFile(goModPath string) (string, []dependency, error) {
	fmt.Printf("Reading go.mod file: %s\n", goModPath)

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
