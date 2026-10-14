package cql

import (
	"fmt"
	"iter"
	"strings"
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

func (e *LexError) Error() string {
	return fmt.Sprintf("%d:%d: %s", e.Position.Line, e.Position.Column, e.Message)
}

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenKeyword
	TokenString
	TokenNumber
	TokenDate
	TokenOperator
	TokenIdent
	TokenLeftParen
	TokenRightParen
	TokenComma
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

func isWordStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isWordPart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.'
}

func isValidKeyword(s string) bool {
	lower := strings.ToLower(s)
	switch lower {
	case "and", "or", "not", "in":
		return true
	default:
		return false
	}
}

func (l *Lexer) lexWord() Token {
	startPos := l.position
	for r, ok := l.peek(); ok && isWordPart(r); r, ok = l.peek() {
		l.advance()
	}

	word := l.input[startPos.Offset:l.position.Offset]
	if isValidKeyword(word) {
		return Token{Type: TokenKeyword, Value: strings.ToLower(word), Position: startPos}
	}
	return Token{Type: TokenIdent, Value: word, Position: startPos}
}

func (l *Lexer) skipDigits() {
	for r, ok := l.peek(); ok && unicode.IsDigit(r); r, ok = l.peek() {
		l.advance()
	}
}

func (l *Lexer) lexNumber() (Token, error) {
	startPos := l.position
	l.skipDigits()
	if r, ok := l.peek(); ok && r == '.' {
		l.advance()
		if r, ok := l.peek(); !ok || !unicode.IsDigit(r) {
			return Token{}, &LexError{Position: startPos, Message: "expected digit after '.'"}
		}
		l.skipDigits()
	}
	return Token{Type: TokenNumber, Value: l.input[startPos.Offset:l.position.Offset], Position: startPos}, nil
}

func isOperator(r rune) bool {
	switch r {
	case '=', '<', '>', '!', '~':
		return true
	default:
		return false
	}
}

func isComplexOperator(s string) bool {
	switch s {
	case "==", "!=", "<=", ">=":
		return true
	default:
		return false
	}
}

func isStringStart(r rune) bool {
	return r == '"' || r == '\''
}

func (l *Lexer) lexString() (Token, error) {
	startPos := l.position
	quote, _ := l.advance()
	var value []rune
	for {
		r, ok := l.peek()
		if !ok {
			return Token{}, &LexError{Position: startPos, Message: "unterminated string"}
		}
		if r == quote {
			l.advance()
			break
		}
		value = append(value, r)
		l.advance()
	}

	return Token{Type: TokenString, Value: string(value), Position: startPos}, nil
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
			case r == ',':
				startPos := l.position
				l.advance()
				token = Token{Type: TokenComma, Value: ",", Position: startPos}
			case r == '(':
				startPos := l.position
				l.advance()
				token = Token{Type: TokenLeftParen, Value: "(", Position: startPos}
			case r == ')':
				startPos := l.position
				l.advance()
				token = Token{Type: TokenRightParen, Value: ")", Position: startPos}
			case unicode.IsSpace(r):
				l.advance()
				continue
			case isOperator(r):
				startPos := l.position
				op := string(r)
				l.advance()
				if nextR, ok := l.peek(); ok && isComplexOperator(op+string(nextR)) {
					op += string(nextR)
					l.advance()
				}

				if op == "!" {
					err = &LexError{Position: startPos, Message: "unsupported operator '!'"}
				}

				if op == "=" {
					token = Token{Type: TokenOperator, Value: "==", Position: startPos}
				} else {
					token = Token{Type: TokenOperator, Value: op, Position: startPos}
				}
			case isStringStart(r):
				token, err = l.lexString()
			case isWordStart(r):
				token = l.lexWord()
			case unicode.IsDigit(r):
				token, err = l.lexNumber()
			default:
				err = &LexError{Position: l.position, Message: fmt.Sprintf("unexpected character %q", r)}
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
