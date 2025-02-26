package utils

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

var LineReplacements = []string{"✔", "- [x]", "☐", "- [ ]"}

type Line struct {
	Content          string
	IndentationDepth int
	IndentationRune  rune
}

func NewLine(content string) *Line {
	line := &Line{}
	// Ensure that we ignore trailing spaces.
	line.Content = strings.TrimRight(content, " ")
	// Decode the first rune, to assert if the line is indented.
	line.IndentationRune, _ = utf8.DecodeRuneInString(line.Content)

	if unicode.IsSpace(line.IndentationRune) {
		line.IndentationDepth = len(line.Content) - len(strings.TrimLeft(line.Content, " "))
		// For non tab space characters, we consider half the depth,
		// since two of those characters should be used per indent.
		if line.IndentationRune != '\t' {
			line.IndentationDepth /= 2
		}

	} else {
		// Reset The indentation rune to avoid missuse.
		line.IndentationRune = 0
	}

	return line
}
