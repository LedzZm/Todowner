package main

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// @TODO: find a better way to represent characters checked for the code to be more readable.
// @TODO: make usable in any directory, regardles of script position.
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

			// Convert todo sections to markdown Headings.
			if strings.HasSuffix(lineContents, ":") {
				lineContents, _ = strings.CutSuffix(lineContents, ":")
				// @todo somehow find subheadings
				lineContents = "# " + lineContents

				editor.WriteString(lineContents + "\n")
				editor.Flush()
				continue
			}

			if lineContents == "" {
				// @todo decide if I should remove empty lines from file,
				// @todo or handle them some other way
				continue
			}

			fixIndentation(&lineContents)

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
		os.Remove(file.Name())
	}
}

// @todo remove abstraction if not needed
func fixIndentation(lineContents *string) {
	_lineContents := *lineContents
	if strings.HasPrefix(_lineContents, "  ") || strings.HasPrefix(_lineContents, "	") {
		// Strip the first indentation.
		// @todo need to decide what to do here.
		_lineContents, result := strings.CutPrefix(_lineContents, "  ")
		if !result {
			_lineContents, result = strings.CutPrefix(_lineContents, "	")
		}

		*lineContents = _lineContents
	}
}
