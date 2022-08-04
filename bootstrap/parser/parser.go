package parser

import (
	"bootstrap/lexer"
	"fmt"
	"strings"
)

func Parse(l *lexer.Lexer) (Ast, bool) {
	err := false
	var lets []*Let
	var funs []*Fun
	var imports []*Import
	for token := l.Peek(); token.Typ != lexer.EOF; token = l.Peek() {
		switch token.Typ {
		case lexer.LET:
			lets = append(lets, parseLet(l))
		case lexer.FUN:
			funs = append(funs, parseFun(l))
		case lexer.IMPORT:
			imports = append(imports, parseImport(l))
		default:
			err = true
			l.ConsumePeek()
		}
	}
	expect(l, lexer.EOF)
	return Ast{Funs: funs, Lets: lets, Imports: imports}, err
}

func expect(l *lexer.Lexer, typ lexer.Typ) lexer.Token {
	token := l.Next()
	if token.Typ != typ {
		panic(fmt.Sprintf("unexpected token '%s' expected '%s'", token.Typ, typ))
	}
	return token
}

func parseImport(l *lexer.Lexer) *Import {
	expect(l, lexer.IMPORT)
	path := parseString(l)
	expect(l, lexer.SEMICOLON)
	return &Import{Path: path}
}

func parseFun(l *lexer.Lexer) *Fun {
	expect(l, lexer.FUN)
	var opts []*Ident
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
	return &Fun{Opts: opts, Ident: ident, Inputs: inputs, Outputs: outputs, Block: block}
}

func parseOpts(l *lexer.Lexer) []*Ident {
	expect(l, lexer.LBRACE)
	idents := parseIdents(l)
	expect(l, lexer.RBRACE)
	return idents
}

func parseIdents(l *lexer.Lexer) []*Ident {
	var idents []*Ident
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

func parseIdent(l *lexer.Lexer) *Ident {
	ident := expect(l, lexer.IDENT)
	return &Ident{Content: ident.Content}
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
	case "void":
		return VOID
	case "string":
		return STRING
	case "char":
		return CHAR
	default:
		panic("unknown type")
	}
}

func parseBlock(l *lexer.Lexer) *Block {
	expect(l, lexer.LBRACE)
	lets := parseLets(l)
	exprs := parseExprs(l)
	expect(l, lexer.RBRACE)
	return &Block{Lets: lets, Exprs: exprs}
}

func parseLets(l *lexer.Lexer) []*Let {
	var lets []*Let
	for {
		if l.Peek().Typ == lexer.LET {
			lets = append(lets, parseLet(l))
		} else {
			return lets
		}
	}
}

func parseLet(l *lexer.Lexer) *Let {
	expect(l, lexer.LET)
	ident := parseIdent(l)
	expect(l, lexer.COLON)
	typ := parseTyp(l)
	exprs := []Expr{}
	if l.Peek().Typ == lexer.EQUALS {
		l.ConsumePeek()
		exprs = parseExprs(l)
	}
	expect(l, lexer.SEMICOLON)
	return &Let{Ident: ident, Typ: typ, Exprs: exprs}
}

func parseExprs(l *lexer.Lexer) []Expr {
	var exprs []Expr
	for {
		peek := l.Peek()
		switch peek.Typ {
		case lexer.IF:
			exprs = append(exprs, parseIf(l))
		case lexer.IDENT:
			exprs = append(exprs, parseIdentExpr(l))
		case lexer.NUMBER:
			exprs = append(exprs, parseNumber(l))
		case lexer.STRING:
			exprs = append(exprs, parseString(l))
		case lexer.CHAR:
			exprs = append(exprs, parseChar(l))
		case lexer.UNWRAP:
			l.ConsumePeek()
			exprs = append(exprs, &Unwrap{})
		default:
			return exprs
		}
	}
}

func parseIf(l *lexer.Lexer) *If {
	expect(l, lexer.IF)
	expect(l, lexer.LPAREN)
	con := parseExprs(l)
	expect(l, lexer.RPAREN)
	exprs := parseExprs(l)
	els := []Expr{}
	if l.Peek().Typ == lexer.ELSE {
		l.ConsumePeek()
		els = parseExprs(l)
	}
	expect(l, lexer.SEMICOLON)
	return &If{Con: con, Exprs: exprs, Else: els}
}

func parseIdentExpr(l *lexer.Lexer) Expr {
	ident := parseIdent(l)
	if l.Peek().Typ == lexer.LPAREN {
		l.ConsumePeek()
		inputs := parseTyps(l)
		expect(l, lexer.COLON)
		outputs := parseTyps(l)
		expect(l, lexer.RPAREN)
		return &Call{Ident: ident, Inputs: inputs, Outputs: outputs}
	} else {
		return ident
	}
}

func parseNumber(l *lexer.Lexer) *Number {
	number := expect(l, lexer.NUMBER)
	var end int
	var typ Typ
	var size int

	base := 10
	start := 0

	if strings.HasPrefix(number.Content, "0x") {
		base = 16
		start = 2
	} else if strings.HasPrefix(number.Content, "0b") {
		base = 2
		start = 2
	}

	if strings.HasSuffix(number.Content, "u8") {
		end = 2
		typ = U8
		size = 1
	} else if strings.HasSuffix(number.Content, "u16") {
		end = 3
		typ = U16
		size = 2
	} else if strings.HasSuffix(number.Content, "u32") {
		end = 3
		typ = U32
		size = 4
	} else if strings.HasSuffix(number.Content, "u64") {
		end = 3
		typ = U64
		size = 8
	} else {
		panic(fmt.Sprintf("number '%s' is missing a type", number.Content))
	}
	content := number.Content[start : len(number.Content)-end]
	return &Number{Content: content, Base: base, Size: size, Typ: typ}
}

func parseString(l *lexer.Lexer) *String {
	str := expect(l, lexer.STRING)
	return &String{Content: str.Content}
}

func parseChar(l *lexer.Lexer) *Char {
	char := expect(l, lexer.CHAR)
	return &Char{Content: char.Content}
}
