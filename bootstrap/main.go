package main

import (
	"bootstrap/lexer"
	"bootstrap/parser"
	"fmt"
)

func main() {
	l := lexer.New(`
	
	fun{inline, asm} nop(:) {
		"nop"
	}
	
	fun{inline, asm} readFile(u64, u64, u64, u64 : u64) {
		"readFile"
	}

	`)
	ast := parser.Parse(l)
	for i := 0; i < len(ast.Funs); i++ {
		fmt.Println(ast.Funs[i])
	}
}
