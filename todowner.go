package main

import (
	"bufio"
	"fmt"
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
		// For non tab space characters, we consider half the depth,
		// since two of those characters should be used per indent.
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
	// Find the .todo files in a given folder, recursively.
	var filePathsToProcess []string
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip hidden directories.
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".todo") {
			filePathsToProcess = append(filePathsToProcess, path)
		}
		return nil
	})

	// Create the backup folder.
	backupDir := "./todowner_backup/"
	if err := os.Mkdir(backupDir, 0777); err != nil {
		// TODO: handle error better than this (?)
		fmt.Println(err.Error())
	}

	for _, filePath := range filePathsToProcess {
		// Initialize the file editor.
		sourceFile, _ := os.Open(filePath)
		defer sourceFile.Close()

		tempFile, _ := os.CreateTemp(
			fmt.Sprint("./", filepath.Dir(sourceFile.Name())),
			filepath.Base(sourceFile.Name()),
		)
		defer tempFile.Close()

		editor := bufio.NewReadWriter(
			bufio.NewReader(sourceFile),
			bufio.NewWriter(tempFile),
		)

		var previousHeadingDepth int
		for {
			_line, _, err := editor.ReadLine()

			if err != nil && err == io.EOF {
				fmt.Println("Finished parsing " + sourceFile.Name())
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

			if !isHeading && Line.indentationDepth <= 0 && !strings.HasPrefix(Line.content, "-") {
				Line.content = Line.content + "⚠️"
			}

			writeLine(*editor, Line.content)
		}

		// markdownFileName := strings.TrimSuffix(sourceFile.Name(), filepath.Ext(sourceFile.Name())) + ".md"

		// os.Rename(tempFile.Name(), markdownFileName)
		// os.MkdirAll(backupDir+sourceFile.Name(), 0777)
	}
}

func writeLine(rw bufio.ReadWriter, s string) {
	rw.WriteString(s + "\n")
	rw.Flush()
}
