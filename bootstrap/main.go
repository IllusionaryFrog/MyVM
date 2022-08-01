package main

import (
	"bootstrap/lexer"
	"fmt"
)

func main() {
	l := lexer.New("test input yep 123 \n fun = ()")

	for token := l.Next(); token.Typ != lexer.EOF; token = l.Next() {
		fmt.Println(token)
	}
}
