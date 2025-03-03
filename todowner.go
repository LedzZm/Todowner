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

// TODO: Try out concurrency... keep current code state binary for benchmarking.
// TODO: Start processing different files to find edge cases.
func main() {

	backupDir := "./todowner_backup/"
	if err := os.Mkdir(backupDir, 0777); err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("Backup created at: %s\n", backupDir)
	}

	for _, filePath := range utils.DiscoverFiles() {
		sourceFile, _ := os.Open(filePath)
		defer sourceFile.Close()

		os.MkdirAll(backupDir+filepath.Dir(filePath), 0770)
		backupFile, _ := os.Create(backupDir + filePath)
		io.Copy(backupFile, sourceFile)
		backupFile.Close()
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

			if strings.HasPrefix(line.Content, "---") {
				line.Content = "--"
				continue
			}

			var isHeading bool
			line.Content, isHeading = strings.CutSuffix(line.Content, ":")
			line.Content = strings.TrimSpace(line.Content)

			if strings.HasPrefix(line.Content, "```") {
				writeLine(*editor, line.Content)
				continue
			}

			if isHeading {
				line.Content = strings.Repeat("#", line.IndentationDepth+1) + " " + line.Content
				writeLine(*editor, line.Content)
				previousHeadingIndentationDepth = line.IndentationDepth
				continue
			}
			// Reduce the nesting of non heading lines..
			newLineIndentation := max(line.IndentationDepth-(previousHeadingIndentationDepth+1), 0)
			// Non tab character spaces, should add twice the depth to the indentation.
			if line.IndentationRune != '	' {
				newLineIndentation *= 2
			}

			// Handle replacements.
			line.Content = strings.NewReplacer(utils.LineReplacements...).Replace(line.Content)
			// Mark for review.
			if !isHeading && !strings.HasPrefix(line.Content, "-") {
				line.Content = "- " + line.Content + "⚠️"
			}

			line.Content = strings.Repeat(string(line.IndentationRune), newLineIndentation) + line.Content
			writeLine(*editor, line.Content)
		}
		// Create the new markdown file.
		markdownFileName := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".md"
		os.Rename(tempFile.Name(), markdownFileName)
		os.Remove(filePath)
	}
}

func writeLine(rw bufio.ReadWriter, s string) {
	rw.WriteString(s + "\n")
	rw.Flush()
}
