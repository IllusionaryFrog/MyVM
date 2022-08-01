package parser

import (
	"bootstrap/lexer"
	"fmt"
)

func Parse(l *lexer.Lexer) Ast {
	var funs []Fun
	for token := l.Peek(); token.Typ != lexer.EOF; token = l.Peek() {
		switch token.Typ {
		case lexer.FUN:
			funs = append(funs, parseFun(l))
		default:
			fmt.Println("error in parsing")
			l.ConsumePeek()
		}
	}
	expect(l, lexer.EOF)
	return Ast{Funs: funs}
}

func expect(l *lexer.Lexer, typ lexer.Typ) lexer.Token {
	token := l.Next()
	if token.Typ != typ {
		fmt.Println(token)
		panic("unexpected token")
	}
	return token
}

func parseFun(l *lexer.Lexer) Fun {
	expect(l, lexer.FUN)
	var opts []Ident
	if l.Peek().Typ == lexer.LBRACE {
		opts = parseOpts(l)
	}
	ident := parseIdent(l)
	expect(l, lexer.LPAREN)
	inputs := parseTyps(l)
	expect(l, lexer.COLON)
	outputs := parseTyps(l)
	expect(l, lexer.RPAREN)
	block := parseBlock(l)
	return Fun{Opts: opts, Ident: ident, Inputs: inputs, Outputs: outputs, Block: block}
}

func parseOpts(l *lexer.Lexer) []Ident {
	expect(l, lexer.LBRACE)
	idents := parseIdents(l)
	expect(l, lexer.RBRACE)
	return idents
}

func parseIdents(l *lexer.Lexer) []Ident {
	var idents []Ident
	for {
		if l.Peek().Typ == lexer.IDENT {
			idents = append(idents, parseIdent(l))
		} else {
			return idents
		}
		if l.Peek().Typ == lexer.COMMA {
			l.ConsumePeek()
		} else {
			return idents
		}
	}
}

func parseIdent(l *lexer.Lexer) Ident {
	ident := expect(l, lexer.IDENT)
	return Ident{Content: ident.Content}
}

func parseTyps(l *lexer.Lexer) []Typ {
	var typs []Typ
	for {
		if l.Peek().Typ == lexer.IDENT {
			typs = append(typs, parseTyp(l))
		} else {
			return typs
		}
		if l.Peek().Typ == lexer.COMMA {
			l.ConsumePeek()
		} else {
			return typs
		}
	}
}

func parseTyp(l *lexer.Lexer) Typ {
	ident := expect(l, lexer.IDENT)
	switch ident.Content {
	case "u8":
		return U8
	case "u16":
		return U16
	case "u32":
		return U32
	case "u64":
		return U64
	case "u128":
		return U128
	case "i8":
		return I8
	case "i16":
		return I16
	case "i32":
		return I32
	case "i64":
		return I64
	case "i128":
		return I128
	default:
		fmt.Println(ident)
		panic("unknown type")
	}
}

func parseBlock(l *lexer.Lexer) Block {
	expect(l, lexer.LBRACE)
	lets := parseLets(l)
	exprs := parseExprs(l)
	expect(l, lexer.RBRACE)
	return Block{Lets: lets, Exprs: exprs}
}

func parseLets(l *lexer.Lexer) []Let {
	var lets []Let
	for {
		if l.Peek().Typ == lexer.LET {
			lets = append(lets, parseLet(l))
		} else {
			return lets
		}
	}
}

func parseLet(l *lexer.Lexer) Let {
	expect(l, lexer.LET)
	ident := parseIdent(l)
	expect(l, lexer.COLON)
	typ := parseTyp(l)
	expect(l, lexer.EQUALS)
	exprs := parseExprs(l)
	expect(l, lexer.SEMICOLON)
	return Let{Ident: ident, Typ: typ, Exprs: exprs}
}

func parseExprs(l *lexer.Lexer) []Expr {
	var exprs []Expr
	for {
		peek := l.Peek()
		switch peek.Typ {
		case lexer.IDENT:
			exprs = append(exprs, parseIdent(l))
		case lexer.NUMBER:
			exprs = append(exprs, parseNumber(l))
		case lexer.STRING:
			exprs = append(exprs, parseString(l))
		case lexer.CHAR:
			exprs = append(exprs, parseChar(l))
		default:
			return exprs
		}
		l.ConsumePeek()
	}
}

func parseNumber(l *lexer.Lexer) Number {
	number := expect(l, lexer.NUMBER)
	return Number{Content: number.Content}
}

func parseString(l *lexer.Lexer) String {
	str := expect(l, lexer.STRING)
	return String{Content: str.Content}
}

func parseChar(l *lexer.Lexer) Char {
	char := expect(l, lexer.CHAR)
	return Char{Content: char.Content}
}
