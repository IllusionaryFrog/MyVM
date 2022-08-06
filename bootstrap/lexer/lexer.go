package lexer

import (
	"fmt"
	"strings"
)

var splitters = [...]string{
	" ", "\t", "\r", "\n",
	":", ";", ",", "\"", "//",
	"(", ")", "{", "}",
}

var num = [...]string{
	"0", "1", "2", "3", "4",
	"5", "6", "7", "8", "9",
}

type Lexer struct {
	cursor int
	input  string
	peeked *Token
}

func New(input string) *Lexer {
	return &Lexer{input: input}
}

func find(s string) int {
	idx := -1
	size := 0
	for i := 0; i < len(splitters); i++ {
		splitter := splitters[i]
		nidx := strings.Index(s, splitter)
		if nidx != -1 && (idx == -1 || nidx < idx) {
			idx = nidx
			size = len(splitter)
		}
	}
	if idx == 0 {
		return size
	}
	return idx
}

func isNumber(s string) bool {
	for i := 0; i < len(num); i++ {
		if strings.HasPrefix(s, num[i]) {
			return true
		}
	}
	return false
}

func (l *Lexer) skipComment() {
	for ; l.cursor < len(l.input); l.cursor++ {
		if l.input[l.cursor] == '\n' {
			return
		}
	}
}

func (l *Lexer) string() string {
	var buf string
	start := l.cursor
	for ; l.cursor < len(l.input); l.cursor++ {
		char := l.input[l.cursor]
		switch char {
		case '"':
			l.cursor++
			return buf + l.input[start:l.cursor-1]
		case '\\':
			if l.cursor+2 >= len(l.input) {
				panic("no end of string")
			}
			buf += l.input[start:l.cursor]
			l.cursor++
			start = l.cursor + 1
			escaped := l.input[l.cursor]
			switch escaped {
			case '\\':
				buf += "\\"
			case 'n':
				buf += "\n"
			case 'r':
				buf += "\r"
			case 't':
				buf += "\t"
			case '"':
				buf += "\""
			default:
				panic(fmt.Sprintf("can't escape '%s' in a string", string(char)))
			}
		default:
		}
	}
	panic("no end of string")
}

func (l *Lexer) ConsumePeek() {
	l.peeked = nil
}

func (l *Lexer) RawNext() Token {
	if l.peeked != nil {
		token := *l.peeked
		l.peeked = nil
		return token
	}

	if l.cursor >= len(l.input) {
		return Token{Typ: EOF}
	}

	var content = l.input[l.cursor:]
	pos := find(content)

	if pos != -1 {
		content = content[:pos]
		l.cursor += pos
	} else {
		l.cursor = len(l.input)
	}

	token := Token{Content: content}

	switch content {
	case " ", "\t", "\r", "\n":
		token.Typ = IGNORED
	case ":":
		token.Typ = COLON
	case ";":
		token.Typ = SEMICOLON
	case ",":
		token.Typ = COMMA
	case "(":
		token.Typ = LPAREN
	case ")":
		token.Typ = RPAREN
	case "{":
		token.Typ = LBRACE
	case "}":
		token.Typ = RBRACE
	case "//":
		token.Typ = IGNORED
		l.skipComment()
	case "\"":
		token.Typ = STRING
		token.Content = l.string()
	case "fun":
		token.Typ = FUN
	case "let":
		token.Typ = LET
	case "if":
		token.Typ = IF
	case "else":
		token.Typ = ELSE
	case "import":
		token.Typ = IMPORT
	case "type":
		token.Typ = TYPE
	case ".unwrap":
		token.Typ = UNWRAP
	case ".wrap":
		token.Typ = WRAP
	case ".addr":
		token.Typ = ADDR
	default:
		if isNumber(content) {
			token.Typ = NUMBER
		} else {
			token.Typ = IDENT
		}
	}

	return token
}

func (l *Lexer) RawPeek() *Token {
	if l.peeked == nil {
		token := l.RawNext()
		l.peeked = &token
	}
	return l.peeked
}

func (l *Lexer) Next() Token {
	var token Token
	for token = l.RawNext(); token.Typ == IGNORED; token = l.RawNext() {
	}
	return token
}

func (l *Lexer) Peek() *Token {
	var token *Token
	for token = l.RawPeek(); token.Typ == IGNORED; token = l.RawPeek() {
		l.ConsumePeek()
	}
	return token
}
