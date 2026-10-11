package cql

import (
	"iter"
	"unicode"
	"unicode/utf8"
)

type Position struct {
	Offset int
	Line   int
	Column int
}

type LexError struct {
	Position Position
	Message  string
}

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenKeyword
	TokenString
	TokenNumber
	TokenDate
	TokenOperator
)

type Token struct {
	Type     TokenType
	Value    string
	Position Position
}

type Lexer struct {
	input    string
	position Position
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:    input,
		position: Position{Offset: 0, Line: 1, Column: 1},
	}
}

func (l *Lexer) peek() (rune, bool) {
	if l.position.Offset >= len(l.input) {
		return 0, false
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.position.Offset:])
	return r, true
}

func (l *Lexer) advance() (rune, bool) {
	if l.position.Offset >= len(l.input) {
		return 0, false
	}
	r, size := utf8.DecodeRuneInString(l.input[l.position.Offset:])
	l.position.Offset += size
	if r == '\n' {
		l.position.Line++
		l.position.Column = 1
	} else {
		l.position.Column++
	}
	return r, true
}

func (l *Lexer) lex() iter.Seq2[Token, error] {
	return func(yield func(Token, error) bool) {
		for l.position.Offset < len(l.input) {
			r, ok := l.peek()
			if !ok {
				break
			}
			
			var token Token
			var err error

			switch {
			case unicode.IsSpace(r):
				l.advance()
				continue 

			}
			if err != nil {
				yield(Token{}, err)
				return
			}
			if !yield(token, nil) {
				return
			}
		}

		yield(Token{Type: TokenEOF, Value: "", Position: l.position}, nil)
	}
}

func (l *Lexer) Stream() *TokenStream {
	return NewTokenStream(l.lex())
}

type TokenStream struct {
	next func() (Token, error, bool)
	stop func()
}

func NewTokenStream(seq iter.Seq2[Token, error]) *TokenStream {
	next, stop := iter.Pull2(seq)
	return &TokenStream{
		next: next,
		stop: stop,
	}
}

func (ts *TokenStream) Next() (Token, error, bool) {
	return ts.next()
}

func (ts *TokenStream) Stop() {
	ts.stop()
}
