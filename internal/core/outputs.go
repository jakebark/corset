package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jakebark/corset/internal/config"
	"github.com/jakebark/corset/internal/inputs"
)

func buildOutput(userInput inputs.UserInput, packedFiles [][]Statement, inputFiles []string) {
	var outputDir string
	if userInput.IsDirectory {
		// For directory replacement, output to the target directory itself
		outputDir = userInput.Target
	} else {
		// For single file replacement, output to the same directory as the input file
		outputDir = filepath.Dir(inputFiles[0])
	}

	if !userInput.IsDirectory && len(inputFiles) == 1 {
		// single file replacement, overwrite
		results := orchestrateOutputFiles(userInput, packedFiles, outputDir, inputFiles)
		reportResults(results)
	} else {
		// directory replacement
		results := orchestrateOutputFiles(userInput, packedFiles, outputDir, inputFiles)
		reportResults(results)
		replaceInputFiles(userInput, inputFiles)
	}
}

func orchestrateOutputFiles(userInput inputs.UserInput, packedFiles [][]Statement, outputDir string, inputFiles []string) []WriteResult {
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
		// single file, use original name
		originalFile := inputFiles[0]
		if fileNum == 1 {
			return originalFile
		}
		// add numeric suffix for splits
		ext := filepath.Ext(originalFile)
		nameWithoutExt := originalFile[:len(originalFile)-len(ext)]
		return fmt.Sprintf("%s-%d%s", nameWithoutExt, fileNum, ext)

	} else if userInput.IsDirectory {
		// use target as base name, add numeric suffix for splits
		baseName := filepath.Base(userInput.Target)
		if fileNum == 1 {
			return filepath.Join(outputDir, baseName+".json")
		}
		return filepath.Join(outputDir, fmt.Sprintf("%s-%d.json", baseName, fileNum))
	}

	// fallback to default naming convention
	return filepath.Join(outputDir, fmt.Sprintf("corset%d.json", fileNum))
}

func writeOutputFile(userInput inputs.UserInput, filename string, statements []Statement) int {
	data := writeJSON(userInput, statements)
	os.WriteFile(filename, data, 0644)
	return len(data)
}

func writeJSON(userInput inputs.UserInput, statements []Statement) []byte {
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

func replaceInputFiles(userInput inputs.UserInput, inputFiles []string) {
	for _, inputFile := range inputFiles {
		os.Remove(inputFile)
	}
}
