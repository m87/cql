package cql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestEmptyInput(t *testing.T) {
	input := ""

	stream := NewLexer(input).Stream()

	tok, err, ok := stream.Next()
	assert.True(t, ok)
	assert.NoError(t, err)
	assert.Equal(t, TokenEOF, tok.Type)
	assert.Equal(t, "", tok.Value)
	assert.Equal(t, Position{Offset: 0, Line: 1, Column: 1}, tok.Position)
}

func TestBlankInput(t *testing.T) {
	input := "   \n  \t  "

	stream := NewLexer(input).Stream()

	tok, err, ok := stream.Next()
	assert.True(t, ok)
	assert.NoError(t, err)
	assert.Equal(t, TokenEOF, tok.Type)
	assert.Equal(t, "", tok.Value)
	assert.Equal(t, Position{Offset: len(input), Line: 2, Column: 6}, tok.Position)
}
