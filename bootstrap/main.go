package main

import (
	"bootstrap/compiler"
	"bootstrap/lexer"
	"bootstrap/parser"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	dat, err := os.ReadFile(os.Args[1])
	base := filepath.Dir(os.Args[1])
	os.Chdir(base)
	if err != nil {
		panic("invalid input")
	}
	l := lexer.New(string(dat))
	ast, perr := parser.Parse(l)
	if perr {
		panic(fmt.Sprintf("error parsing file '%s'", os.Args[1]))
	}
	bytes := compiler.Compile(ast)
	if os.WriteFile(os.Args[2], bytes, fs.FileMode(os.O_TRUNC|os.O_CREATE|os.O_RDWR)) != nil {
		panic("invalid output")
	}
}
