package cql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newStream(t *testing.T, input string) *TokenStream {
	t.Helper()
	stream := NewLexer(input).Stream()
	t.Cleanup(stream.Stop)
	return stream
}

func requireNext(t *testing.T, stream *TokenStream) Token {
	t.Helper()
	tok, err, ok := stream.Next()
	require.True(t, ok)
	require.NoError(t, err)
	return tok
}

func requireEOF(t *testing.T, stream *TokenStream, pos Position) {
	t.Helper()
	tok := requireNext(t, stream)
	assert.Equal(t, TokenEOF, tok.Type)
	assert.Equal(t, "", tok.Value)
	assert.Equal(t, pos, tok.Position)
}

func TestEmptyInput(t *testing.T) {
	stream := newStream(t, "")

	requireEOF(t, stream, Position{Offset: 0, Line: 1, Column: 1})
}

func TestBlankInput(t *testing.T) {
	input := "   \n  \t  "
	stream := newStream(t, input)

	requireEOF(t, stream, Position{Offset: len(input), Line: 2, Column: 6})
}

func TestSingleRuneOperator(t *testing.T) {
	inputs := []string{"+", "-", "*", "/", "=", "<", ">"}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			stream := newStream(t, input)

			tok := requireNext(t, stream)
			assert.Equal(t, TokenOperator, tok.Type)
			assert.Equal(t, input, tok.Value)
			assert.Equal(t, Position{Offset: 0, Line: 1, Column: 1}, tok.Position)

			requireEOF(t, stream, Position{Offset: len(input), Line: 1, Column: len(input) + 1})
		})
	}
}

func TestComplexOperator(t *testing.T) {
	inputs := []string{"==", "!=", "<=", ">="}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			stream := newStream(t, input)

			tok := requireNext(t, stream)
			assert.Equal(t, TokenOperator, tok.Type)
			assert.Equal(t, input, tok.Value)
			assert.Equal(t, Position{Offset: 0, Line: 1, Column: 1}, tok.Position)

			requireEOF(t, stream, Position{Offset: len(input), Line: 1, Column: len(input) + 1})
		})
	}
}

func TestStringLiteral(t *testing.T) {
	input := `"Hello, World!"`
	stream := newStream(t, input)

	tok := requireNext(t, stream)
	assert.Equal(t, TokenString, tok.Type)
	assert.Equal(t, "Hello, World!", tok.Value)
	assert.Equal(t, Position{Offset: 0, Line: 1, Column: 1}, tok.Position)

	requireEOF(t, stream, Position{Offset: len(input), Line: 1, Column: len(input) + 1})
}

func TestUnterminatedString(t *testing.T) {
	stream := newStream(t, `<= "abc`)

	requireNext(t, stream)

	_, err, ok := stream.Next()
	require.True(t, ok)
	var lexErr *LexError
	require.ErrorAs(t, err, &lexErr)
	assert.Equal(t, Position{Offset: 3, Line: 1, Column: 4}, lexErr.Position)

	_, _, ok = stream.Next()
	assert.False(t, ok)
}

func TestMultipleTokens(t *testing.T) {
	input := "<= 'a'\n!= \"b\""
	stream := newStream(t, input)

	expected := []Token{
		{Type: TokenOperator, Value: "<=", Position: Position{Offset: 0, Line: 1, Column: 1}},
		{Type: TokenString, Value: "a", Position: Position{Offset: 3, Line: 1, Column: 4}},
		{Type: TokenOperator, Value: "!=", Position: Position{Offset: 7, Line: 2, Column: 1}},
		{Type: TokenString, Value: "b", Position: Position{Offset: 10, Line: 2, Column: 4}},
	}
	for _, want := range expected {
		assert.Equal(t, want, requireNext(t, stream))
	}

	requireEOF(t, stream, Position{Offset: len(input), Line: 2, Column: 7})
}

func TestKeywordToken(t *testing.T) {
	input := "aNd OR not"
	stream := newStream(t, input)

	expected := []Token{
		{Type: TokenKeyword, Value: "and", Position: Position{Offset: 0, Line: 1, Column: 1}},
		{Type: TokenKeyword, Value: "or", Position: Position{Offset: 4, Line: 1, Column: 5}},
		{Type: TokenKeyword, Value: "not", Position: Position{Offset: 7, Line: 1, Column: 8}},
	}
	for _, want := range expected {
		assert.Equal(t, want, requireNext(t, stream))
	}

	requireEOF(t, stream, Position{Offset: len(input), Line: 1, Column: len(input) + 1})
}

func TestComplexInput(t *testing.T) {
	input := `AND "string" <= 123.45 OR 'another string' != 'string'`
	stream := newStream(t, input)

	expected := []Token{
		{Type: TokenKeyword, Value: "and", Position: Position{Offset: 0, Line: 1, Column: 1}},
		{Type: TokenString, Value: "string", Position: Position{Offset: 4, Line: 1, Column: 5}},
		{Type: TokenOperator, Value: "<=", Position: Position{Offset: 13, Line: 1, Column: 14}},
		{Type: TokenNumber, Value: "123.45", Position: Position{Offset: 16, Line: 1, Column: 17}},
		{Type: TokenKeyword, Value: "or", Position: Position{Offset: 23, Line: 1, Column: 24}},
		{Type: TokenString, Value: "another string", Position: Position{Offset: 26, Line: 1, Column: 27}},
		{Type: TokenOperator, Value: "!=", Position: Position{Offset: 43, Line: 1, Column: 44}},
		{Type: TokenString, Value: "string", Position: Position{Offset: 46, Line: 1, Column: 47}},
	}
	for _, want := range expected {
		assert.Equal(t, want, requireNext(t, stream))
	}
	requireEOF(t, stream, Position{Offset: len(input), Line: 1, Column: len(input) + 1})
}

func TestIdentToken(t *testing.T) {
	input := "order_id android _x1"
	stream := newStream(t, input)

	expected := []Token{
		{Type: TokenIdent, Value: "order_id", Position: Position{Offset: 0, Line: 1, Column: 1}},
		{Type: TokenIdent, Value: "android", Position: Position{Offset: 9, Line: 1, Column: 10}},
		{Type: TokenIdent, Value: "_x1", Position: Position{Offset: 17, Line: 1, Column: 18}},
	}
	for _, want := range expected {
		assert.Equal(t, want, requireNext(t, stream))
	}

	requireEOF(t, stream, Position{Offset: len(input), Line: 1, Column: len(input) + 1})
}

func TestNumberToken(t *testing.T) {
	inputs := []string{"0", "42", "123.45"}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			stream := newStream(t, input)

			tok := requireNext(t, stream)
			assert.Equal(t, Token{Type: TokenNumber, Value: input, Position: Position{Offset: 0, Line: 1, Column: 1}}, tok)

			requireEOF(t, stream, Position{Offset: len(input), Line: 1, Column: len(input) + 1})
		})
	}
}

func TestParentheses(t *testing.T) {
	input := "(a AND b) OR (c != d)"
	stream := newStream(t, input)

	expected := []Token{
		{Type: TokenLeftParen, Value: "(", Position: Position{Offset: 0, Line: 1, Column: 1}},
		{Type: TokenIdent, Value: "a", Position: Position{Offset: 1, Line: 1, Column: 2}},
		{Type: TokenKeyword, Value: "and", Position: Position{Offset: 3, Line: 1, Column: 4}},
		{Type: TokenIdent, Value: "b", Position: Position{Offset: 7, Line: 1, Column: 8}},
		{Type: TokenRightParen, Value: ")", Position: Position{Offset: 8, Line: 1, Column: 9}},
		{Type: TokenKeyword, Value: "or", Position: Position{Offset: 10, Line: 1, Column: 11}},
		{Type: TokenLeftParen, Value: "(", Position: Position{Offset: 13, Line: 1, Column: 14}},
		{Type: TokenIdent, Value: "c", Position: Position{Offset: 14, Line: 1, Column: 15}},
		{Type: TokenOperator, Value: "!=", Position: Position{Offset: 16, Line: 1, Column: 17}},
		{Type: TokenIdent, Value: "d", Position: Position{Offset: 19, Line: 1, Column: 20}},
		{Type: TokenRightParen, Value: ")", Position: Position{Offset: 20, Line: 1, Column: 21}},
	}
	for _, want := range expected {
		assert.Equal(t, want, requireNext(t, stream))
	}

	requireEOF(t, stream, Position{Offset: len(input), Line: 1, Column: len(input) + 1})
}
