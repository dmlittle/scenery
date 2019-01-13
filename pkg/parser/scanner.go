package parser

import (
	"fmt"
	"io"
	"strconv"
	"text/scanner"
	"unicode/utf8"

	"github.com/alecthomas/participle/lexer"
)

// SceneryDefinition is a copy of participles default Definition lexer with the
// modification that new lines are not ignored.
type SceneryDefinition struct{}

// textScannerLexer is a Lexer based on text/scanner.Scanner
type textScannerLexer struct {
	scanner  *scanner.Scanner
	filename string
}

// Lex an io.Reader with text/scanner.Scanner.
//
// This provides very fast lexing of source code compatible with Go tokens.
//
// Note that this differs from text/scanner.Scanner in that string tokens will
// be unquoted and new lines are not ignored.
func (d *SceneryDefinition) Lex(r io.Reader) (lexer.Lexer, error) {
	l := &textScannerLexer{
		filename: nameOfReader(r),
		scanner:  &scanner.Scanner{},
	}
	l.scanner.Init(r)
	l.scanner.Whitespace ^= 1 << '\n' // don't skip new lines
	l.scanner.Error = func(s *scanner.Scanner, msg string) {
		// This is to support single quoted strings. Hacky.
		if msg != "illegal char literal" {
			panic(lexer.Errorf(lexer.Position(l.scanner.Pos()), msg))
		}
	}
	return l, nil
}

// Symbols returns a map of the tokens type symbols
func (d *SceneryDefinition) Symbols() map[string]rune {
	return map[string]rune{
		"EOF":       scanner.EOF,
		"Char":      scanner.Char,
		"Ident":     scanner.Ident,
		"Int":       scanner.Int,
		"Float":     scanner.Float,
		"String":    scanner.String,
		"RawString": scanner.RawString,
		"Comment":   scanner.Comment,
	}
}

func (t *textScannerLexer) Next() (lexer.Token, error) {
	typ := t.scanner.Scan()
	text := t.scanner.TokenText()
	pos := lexer.Position(t.scanner.Position)
	pos.Filename = t.filename
	return textScannerTransform(lexer.Token{
		Type:  typ,
		Value: text,
		Pos:   pos,
	})
}

func textScannerTransform(token lexer.Token) (lexer.Token, error) {
	// Unquote strings.
	switch token.Type {
	case scanner.Char:
		// FIXME(alec): This is pretty hacky...we convert a single quoted char into a double
		// quoted string in order to support single quoted strings.
		token.Value = fmt.Sprintf("\"%s\"", token.Value[1:len(token.Value)-1])
		fallthrough
	case scanner.String:
		s, err := strconv.Unquote(token.Value)
		if err != nil {
			return lexer.Token{}, lexer.Errorf(token.Pos, "%s: %q", err.Error(), token.Value)
		}
		token.Value = s
		if token.Type == scanner.Char && utf8.RuneCountInString(s) > 1 {
			token.Type = scanner.String
		}
	case scanner.RawString:
		token.Value = token.Value[1 : len(token.Value)-1]
	}
	return token, nil
}

func nameOfReader(r interface{}) string {
	if nr, ok := r.(interface{ Name() string }); ok {
		return nr.Name()
	}
	return ""
}
