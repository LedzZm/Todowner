package main

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Line struct {
	content          string
	indentationDepth int
	indentationRune  rune
}

func NewLine(content string) *Line {
	Line := &Line{}
	// Ensure that we ignore trailing spaces.
	Line.content = strings.TrimRight(content, " ")
	// Decode the first rune, to assert if the line is indented.
	Line.indentationRune, _ = utf8.DecodeRuneInString(Line.content)

	if unicode.IsSpace(Line.indentationRune) {
		Line.indentationDepth = len(Line.content) - len(strings.TrimLeft(Line.content, " "))

		if Line.indentationRune != '\t' {
			Line.indentationDepth /= 2
		}

	} else {
		// Reset The indentation rune to avoid missuse.
		Line.indentationRune = 0
	}

	return Line
}

func main() {
	files, _ := os.ReadDir(".")

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".todo") {
			continue
		}
		// Initialize the file editor.
		sourceFile, _ := os.Open(file.Name())
		defer sourceFile.Close()
		markdownFileName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())) + ".md"
		// @TODO: TBD, should I use "" for the default tmp dir of the system when binary?
		tempFile, _ := os.CreateTemp(".", markdownFileName)
		defer tempFile.Close()
		editor := bufio.NewReadWriter(
			bufio.NewReader(sourceFile),
			bufio.NewWriter(tempFile),
		)

		var previousHeadingDepth int
		for {
			_line, _, err := editor.ReadLine()

			if err != nil && err == io.EOF {
				println("Finished parsing " + file.Name())
				break
			}

			Line := *NewLine(string(_line[:]))
			if Line.content == "" {
				continue
			}
			// Convert todo sections to markdown Headings.
			var isHeading bool
			Line.content, isHeading = strings.CutSuffix(Line.content, ":")
			if isHeading {
				// Headings should have no indentation.
				Line.content = strings.TrimLeft(Line.content, string(Line.indentationRune))
				Line.content = strings.Repeat("#", Line.indentationDepth+1) + " " + Line.content
				// Store the heading depth, for processing the content under it.
				previousHeadingDepth = Line.indentationDepth
				writeLine(*editor, Line.content)
				continue
			}
			// If the indentation is not the tab character,
			// consider the indentation to be double the amount runes.
			if Line.indentationRune != '	' {
				previousHeadingDepth *= 2
			}
			for range previousHeadingDepth {
				Line.content = strings.TrimPrefix(Line.content, string(Line.indentationRune))
			}
			// Convert the todo boxes to markdown checkboxes.
			Line.content = strings.ReplaceAll(Line.content, "☐", "- [ ]")

			writeLine(*editor, Line.content)
		}

		os.Rename(tempFile.Name(), markdownFileName)
		os.Remove(file.Name())
	}
}

func writeLine(rw bufio.ReadWriter, s string) {
	rw.WriteString(s + "\n")
	rw.Flush()
}
