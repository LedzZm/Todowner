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

// @TODO: Find a better way to represent characters checked for the code to be more readable.
// @TODO: Make usable in any directory, regardles of script position.
// @TODO: Fix comments (Document properly and correct comment formatting).
func main() {
	files, _ := os.ReadDir(".")

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".todo") {
			continue
		}

		// @TODO: Do I need error handling here?
		// Initialize the file editor.
		sourceFile, _ := os.Open(file.Name())
		defer sourceFile.Close()
		markdownFileName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())) + ".md"
		// @TODO: TBD, should I use "" for the default tmp dir of the system when binary?
		tempFile, _ := os.CreateTemp(".", markdownFileName)
		// @TODO: is this even needed?
		// @TODO: Maybe with concurrency
		defer tempFile.Close()
		editor := bufio.NewReadWriter(
			bufio.NewReader(sourceFile),
			bufio.NewWriter(tempFile),
		)

		lineNumber := 1
		for {
			_line, _, err := editor.ReadLine()

			if err != nil && err == io.EOF {
				println("Finished parsing " + file.Name())
				break
			}

			lineContents := string(_line[:])
			prefix, hasIndentation := resolveIndentation(lineContents)

			var depth int = 0
			if hasIndentation {
				// TODO: depth%2 might be off if " " is the indentation character
				depth = findIndentationDepth(lineContents, prefix)
				koromposIndentations(&lineContents)
			}

			// Convert todo sections to markdown Headings.
			lineContents, isHeading := strings.CutSuffix(lineContents, ":")
			if isHeading {
				lineContents = strings.Repeat("#", depth+1) + " " + lineContents

				editor.WriteString(lineContents + "\n")
				editor.Flush()
				continue
			}

			if lineContents == "" {
				// @todo decide if I should remove empty lines from file,
				// @todo or handle them some other way
				continue
			}

			// @TODO: Is this the best place to do it here, or for every line.
			// @TODO: Reconsider `Is this the best place to do it here, or for every line.`
			//     When processing the files concurrently
			editor.WriteString(lineContents + "\n")
			editor.Flush()

			// result := strings.ReplaceAll(lineContents, "☐", "- [ ]")

			// strings.Split(result, "")
			// fmt.Println(result)
			lineNumber++
		}

		// @TODO: Can I do this with one operation? Not needed just flex.
		os.Rename(tempFile.Name(), markdownFileName)
		// os.Remove(file.Name())
	}
}

// @todo remove abstraction if not needed
func koromposIndentations(lineContents *string) {
	_lineContents := *lineContents
	prefix, hasIndentation := resolveIndentation(_lineContents)

	if hasIndentation {
		*lineContents = strings.TrimPrefix(_lineContents, string(prefix))
	}
}

func findIndentationDepth(haystack string, target string) int {

	count := 0
	for _, char := range haystack {
		// Stop counting once the target character stops appearing.
		if string(char) != target {
			return count
		}
		count++
	}

	// https://www.practical-go-lessons.com/chap-34-benchmarks
	// TODO: len(haystack)-len(strings.TrimLeft(haystack, " "))
	return count
}

// @TODO: Need something more generic for the runes?
func resolveIndentation(line string) (string, bool) {
	firstRune, _ := utf8.DecodeRuneInString(line)
	prefix := string(firstRune)
	if firstRune == ' ' {

		c := findIndentationDepth(line, " ")

		if c%2 != 0 {
			// println(c, c2)
		}

		// println(findIndentationDepth(line, " "))

		prefix = "  "
	}

	unicode.IsSpace(firstRune)

	// if prefix == " " {
	// 	depth := findIndentationDepth(line, prefix)
	// 	return "  ", true
	// }

	return prefix, string(prefix) == "  " || prefix == "	"
}
