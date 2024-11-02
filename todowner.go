package main

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// @TODO: Find a better way to represent characters checked for the code to be more readable.
// @TODO: Make usable in any directory, regardles of script position.
// @TODO: Fix comments (Document properly and correct comment formatting).
// @TODO: Try moving indentation handling to another packate @optional.
func main() {
	files, _ := os.ReadDir(".")

	println("Decide what to do with double spaces.\nDecide if I want to remove abstractions. \nPass the indentation and prefix to koromposIndentations (Decide if I want to pass double space as string or as rune to handle)")
	os.Exit(5)

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

			// Read the line and strip suffixing spaces.
			lineContents := strings.TrimRight(string(_line[:]), " ")
			prefix, hasIndentation := resolveIndentation(lineContents)

			var depth int = 0
			if hasIndentation {
				// TODO: depth%2 might be off if " " is the indentation character
				depth = findIndentationDepth(lineContents, prefix)
				koromposIndentations(&lineContents)
			}

			println("depth", depth)

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

// TODO: When finished with usages, evaluate if needs some type so stcache
func findIndentationDepth(haystack string, target string) int {
	count := len(haystack) - len(strings.TrimLeft(haystack, " "))

	if target == " " {
		count /= 2
	}

	return count
}

// @TODO: Need something more generic for the runes?
// @TODO: Remove this abstraction?
func resolveIndentation(line string) (string, bool) {
	prefix, _ := utf8.DecodeRuneInString(line)
	return string(prefix), prefix == ' ' || prefix == '	'
}
