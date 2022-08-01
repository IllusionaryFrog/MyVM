package main

import (
	"bootstrap/compiler"
	"bootstrap/lexer"
	"bootstrap/parser"
	"io/fs"
	"os"
)

func main() {
	dat, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic("invalid input")
	}
	l := lexer.New(string(dat))
	ast := parser.Parse(l)
	bytes := compiler.Compile(ast)
	if os.WriteFile(os.Args[2], bytes, fs.FileMode(os.O_TRUNC|os.O_CREATE|os.O_RDWR)) != nil {
		panic("invalid output")
	}
}
