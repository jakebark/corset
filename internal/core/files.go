package core

import (
	"io/fs"
	"path/filepath"
	"strings"
)

func FindJSONFilesInDirectory(dir string) []string {
	var jsonFiles []string
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			jsonFiles = append(jsonFiles, path)
		}
		return nil
	})
	return jsonFiles
}
