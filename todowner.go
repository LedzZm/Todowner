package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

// @todo find a better way to represent characters checked for the code to be more readable.
// @make usable anywhere.
// When writing, use *_converted.md for testing environment.

func main() {
	// @todo add comment when I have decided what this will be.
	const (
		heading   = 0 // seek relative to the origin of the file
		korompos  = 1 // seek relative to the current offset
		endrompos = 2 // seek relative to the end
	)

	files, _ := os.ReadDir(".")

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".todo") {
			continue
		}

		stream, _ := os.OpenFile(file.Name(), os.O_RDWR, 0755)

		reader := bufio.NewReader(stream)
		writer := bufio.NewWriter(stream)
		editor := bufio.NewReadWriter(reader, writer)

		lineNumber := 1
		var previousLine byte
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

				// @todo Abstract?
				previousLine = heading
				continue
			}

			if previousLine == heading {
				if slices.Contains([]string{"  ", "	"}, lineContents) {
					fmt.Println(lineContents)
					os.Exit(0)
				} else {
					// @todo add - in the beginning
					// @todo do not do this in else body.
					continue
				}

			}

			os.Exit(0)
			if string(lineContents[0]) != "\t" {
				continue
			}

			fmt.Println(string(lineContents[0]))

			os.Exit(1)

			// result := strings.ReplaceAll(lineContents, "☐", "- [ ]")

			// strings.Split(result, "")
			// fmt.Println(result)
			lineNumber++

			os.Exit(2)
		}

	}

}
