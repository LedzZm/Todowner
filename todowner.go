package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	utils "todowner/src"
)

// @TODO: Prompt user for file or directory.
// @TODO: If passed file directly do not walk directory
// @TODO: Add progressbar
// @TODO: Recursive should be optional (?) -r
// @TODO: Create doc file to clean up the main function.
// @TODO: Code splitting https://github.com/golang-standards/project-layout
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
		// Copy the sourceFile Contents to the backu file.
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

		var previousHeadingIndentationDepth int
		for {
			_line, _, err := editor.ReadLine()

			if err != nil && err == io.EOF {
				fmt.Println("Finished parsing " + filePath)
				break
			}

			line := *utils.NewLine(string(_line[:]))
			if line.Content == "" {
				continue
			}

			if line.Content == "＿＿＿＿＿＿＿＿＿＿＿＿＿＿＿＿＿＿＿" {
				line.Content = "--"
				continue
			}

			// Convert todo sections to markdown Headings.
			var isHeading bool
			line.Content, isHeading = strings.CutSuffix(line.Content, ":")
			// Start by trimming spaces, to remove indentation.
			line.Content = strings.TrimSpace(line.Content)
			// Leave code blocks as is.
			if strings.HasPrefix(line.Content, "```") {
				writeLine(*editor, line.Content)
				continue
			}
			// Prefix headings with the appropriate number of # characters,
			// based on the indentation depth.
			if isHeading {
				line.Content = strings.Repeat("#", line.IndentationDepth+1) + " " + line.Content
				writeLine(*editor, line.Content)
				// At the end of processing the line, store the current indentation depth,
				// to know if the next line should be indented and how much.
				previousHeadingIndentationDepth = line.IndentationDepth
				continue
			}
			// Non-heading lines should have their nesting levels reduced
			// by the number of levels lost by the containing heading.
			// The old format was based on the premise that:
			// 1. Headings end with ':'
			// 2. Content starts one indentation level after the heading.
			// In markdown this is not needed, so we also remove one more indentation level.
			// in non-heading lines.
			newLineIndentation := max(line.IndentationDepth-(previousHeadingIndentationDepth+1), 0)
			// Non tab character spaces, should add twice the depth to the indentation.
			if line.IndentationRune != '	' {
				newLineIndentation *= 2
			}

			// Convert the todo boxes to markdown checkboxes.
			line.Content = strings.NewReplacer(utils.LineReplacements...).Replace(line.Content)

			// Add a warning to lines that are not headings and do not start with a dash.
			// This will pre-emptively mark them as list items, but also notify
			// the user that they might need to be reviewed.
			if !isHeading && !strings.HasPrefix(line.Content, "-") {
				line.Content = "- " + line.Content + "⚠️"
			}

			line.Content = strings.Repeat(string(line.IndentationRune), newLineIndentation) + line.Content

			writeLine(*editor, line.Content)
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
