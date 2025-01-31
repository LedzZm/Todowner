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

// @TODO: Prompt user for file or directory.
// @TODO: If passed file directly do not walk directory
// @TODO: Add progressbar
// @TODO: Recursive should be optional (?) -r
// @TODO: Create doc file to clean up the main function.
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
	} else {
		fmt.Printf("Backup created at: %s\n", backupDir)
	}

	for _, filePath := range filePathsToProcess {
		sourceFile, _ := os.Open(filePath)
		defer sourceFile.Close()

		// Create the full nested filepath inside the backup folder.
		os.MkdirAll(backupDir+filepath.Dir(filePath), 0770)
		// Copy the sourceFile contents to the backu file.
		// backupFile, _ := os.Open(backupDir + filePath)
		backupFile, _ := os.Create(backupDir + filePath)
		io.Copy(backupFile, sourceFile)
		backupFile.Close()
		// Reset the sourceFile pointer to the start of the file.
		sourceFile.Seek(0, 0)

		tempFile, _ := os.CreateTemp(
			fmt.Sprint("./", filepath.Dir(filePath)),
			filepath.Base(filePath),
		)
		defer tempFile.Close()

		editor := bufio.NewReadWriter(
			bufio.NewReader(sourceFile),
			bufio.NewWriter(tempFile),
		)

		var previousLineIndentationDepth int
		for {
			_line, _, err := editor.ReadLine()

			if err != nil && err == io.EOF {
				fmt.Println("Finished parsing " + filePath)
				break
			}

			Line := *NewLine(string(_line[:]))
			if Line.content == "" {
				continue
			}
			// Convert todo sections to markdown Headings.
			var isHeading bool
			Line.content, isHeading = strings.CutSuffix(Line.content, ":")
			// Start by trimming spaces, to remove indentation.
			Line.content = strings.TrimSpace(Line.content)
			// Leave code blocks as is.
			if strings.HasPrefix(Line.content, "```") {
				writeLine(*editor, Line.content)
				continue
			}
			// Prefix headings with the appropriate number of # characters,
			// based on the indentation depth.
			if isHeading {
				Line.content = strings.Repeat("#", Line.indentationDepth+1) + " " + Line.content
				writeLine(*editor, Line.content)
				// Reset previous line depth, since we are in a new heading.
				// and the next line should not be indented.
				previousLineIndentationDepth = 0
				continue
			}
			// Non heading lines should be indented based on the previous line.
			// If the indentation is not the tab character,
			// consider the indentation to be double the amount runes.
			if Line.indentationDepth != 0 && Line.indentationRune != 0 && previousLineIndentationDepth != 0 && previousLineIndentationDepth != Line.indentationDepth {
				if Line.indentationRune != '	' {
					previousLineIndentationDepth *= 2
				}
				Line.content = strings.Repeat(string(Line.indentationRune), previousLineIndentationDepth) + Line.content
			}

			// for range previousLineIndentationDepth {
			// 	Line.content = string(Line.indentationRune) + Line.content
			// }

			// Convert the todo boxes to markdown checkboxes.
			Line.content = strings.NewReplacer("✔", "- [x]", "☐", "- [ ]").Replace(Line.content)

			// @TODO: Do I actually need this?
			if !isHeading && Line.indentationDepth <= 0 && !strings.HasPrefix(Line.content, "-") {
				Line.content = Line.content + "⚠️"
			}
			// At the end of processing the line, store the current indentation depth,
			// to know if the next line should be indented and how much.
			fmt.Println("The next line should be indented by the difference of this line and previous line depth")
			fmt.Println("Should consider if the next line is more or less indented than the previous line")
			fmt.Println("1.1, 1.2, 1.3, 1.3.1, 1.3.2, 1.3.3, 1.4, 1.5")
			fmt.Println("In the above sequence, 1.4 should not take 1.3.3 into consideration but 1.3")

			return
			if previousLineIndentationDepth != Line.indentationDepth {
				previousLineIndentationDepth = max(Line.indentationDepth-previousLineIndentationDepth, 0)
			}

			writeLine(*editor, Line.content)
		}
		// Create the new markdown file.
		markdownFileName := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".md"
		os.Rename(tempFile.Name(), markdownFileName)
	}
}

func writeLine(rw bufio.ReadWriter, s string) {
	rw.WriteString(s + "\n")
	rw.Flush()
}
