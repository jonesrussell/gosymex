# GoSymEx: Golang Symbol Extractor

## Overview
GoSymEx is a set of commands designed to make it easier to code with chatbots. It describes Go source code files and extracts only the relevant information, making it easier to understand large and complex files.

## Features
- Makes it easier to pair code with chatbots by providing a simplified view of Go code.
- Describes Go files and returns their abstract syntax tree (AST).
- Extracts relevant information from different types of AST nodes.

## Usage
To describe a Go file, use the `describe` command followed by the path to the file. This will print out details about the file in JSON format.

Command: `gosymex describe <filepath>`

## Installation
To install the program, clone this repository and build the program using Go’s built-in toolchain. For example:

    git clone https://github.com/jonesrussell/gosymex.git
    cd gosymex
    go build .

## Contributing
Contributions are welcome! To contribute to this project, you can follow these steps:

1. Fork the repository.
2. Create a new branch for your changes.
3. Make your changes in your branch.
