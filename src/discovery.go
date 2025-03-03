package utils

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
)

func DiscoverFiles() []string {
	var filePathsToProcess []string
	filePathFlag := flag.String("f", "", "Specify a .todo file to process")
	flag.Parse()

	if *filePathFlag != "" {
		filePathsToProcess = append(filePathsToProcess, *filePathFlag)
		return filePathsToProcess
	}

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

	return filePathsToProcess
}
