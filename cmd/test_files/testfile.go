package main

import (
	"fmt"
	"net/http"
)

type MyStruct struct {
	Field1 int
	Field2 string
}

func MyFunc(param1 int, param2 string) (result bool) {
	fmt.Println("Hello, world!")
	return true
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, world!")
	})
	http.ListenAndServe(":8080", nil)
}
