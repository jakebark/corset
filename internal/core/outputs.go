package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jakebark/corset/internal/config"
	"github.com/jakebark/corset/internal/inputs"
)

func writeOutputFiles(userInput inputs.UserInput, packedFiles [][]Statement, inputFiles []string) {
	outputDir := filepath.Dir(inputFiles[0])
	
	// For single file replacement, we need special handling to avoid deleting before writing
	if !userInput.IsDirectory && len(inputFiles) == 1 {
		// Single file replacement: write to the same file (overwrite)
		results := writeAllPolicyFiles(userInput, packedFiles, outputDir, inputFiles)
		reportResults(results)
		// No need to delete - we overwrote the file
	} else {
		// Directory replacement
		results := writeAllPolicyFiles(userInput, packedFiles, outputDir, inputFiles)
		reportResults(results)
		replaceInputFiles(userInput, inputFiles)
	}
}

func writeAllPolicyFiles(userInput inputs.UserInput, packedFiles [][]Statement, outputDir string, inputFiles []string) []WriteResult {
	var results []WriteResult
	for i, statements := range packedFiles {
		filename := generateOutputFilename(userInput, outputDir, i+1, inputFiles)
		size := writeOutputFile(userInput, filename, statements)
		results = append(results, WriteResult{
			Filename:   filename,
			Size:       size,
			Statements: len(statements),
		})
	}
	return results
}

func generateOutputFilename(userInput inputs.UserInput, outputDir string, fileNum int, inputFiles []string) string {
	if !userInput.IsDirectory && len(inputFiles) == 1 {
		// Single file replacement - use original filename
		return inputFiles[0]
	} else if userInput.IsDirectory {
		// Directory replacement - use target as base name with number suffix
		baseName := filepath.Base(userInput.Target)
		if fileNum == 1 {
			return filepath.Join(outputDir, baseName+".json")
		}
		return filepath.Join(outputDir, fmt.Sprintf("%s-%d.json", baseName, fileNum))
	}
	
	// Fallback to default corset naming convention
	return filepath.Join(outputDir, fmt.Sprintf("corset%d.json", fileNum))
}

func createPolicyJSON(userInput inputs.UserInput, statements []Statement) []byte {
	policy := Policy{
		Version:   config.SCPVersion,
		Statement: make([]map[string]interface{}, len(statements)),
	}

	for i, stmt := range statements {
		policy.Statement[i] = stmt.Content
	}

	if userInput.Whitespace {
		data, _ := json.MarshalIndent(policy, "", "  ")
		return data
	}
	data, _ := json.Marshal(policy)
	return data
}

func reportResults(results []WriteResult) {
	fmt.Printf("Split into %d files:\n", len(results))
	for _, result := range results {
		fmt.Printf("- %s (%d characters, %d statements)\n",
			filepath.Base(result.Filename), result.Size, result.Statements)
	}
}

func writeOutputFile(userInput inputs.UserInput, filename string, statements []Statement) int {
	data := createPolicyJSON(userInput, statements)
	os.WriteFile(filename, data, 0644)
	return len(data)
}

func replaceInputFiles(userInput inputs.UserInput, inputFiles []string) {
	// Always replace/delete input files since replacement is now automatic
	for _, inputFile := range inputFiles {
		os.Remove(inputFile)
	}
}
