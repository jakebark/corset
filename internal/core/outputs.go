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
	results := writeAllPolicyFiles(userInput, packedFiles, outputDir)
	reportResults(results)
	deleteInputFiles(userInput, inputFiles)
}

func writeAllPolicyFiles(userInput inputs.UserInput, packedFiles [][]Statement, outputDir string) []WriteResult {
	var results []WriteResult
	for i, statements := range packedFiles {
		filename := filepath.Join(outputDir, fmt.Sprintf("corset%d.json", i+1))
		size := writeOutputFile(userInput, filename, statements)
		results = append(results, WriteResult{
			Filename:   filename,
			Size:       size,
			Statements: len(statements),
		})
	}
	return results
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

func deleteInputFiles(userInput inputs.UserInput, inputFiles []string) {
	if userInput.Delete {
		if userInput.IsDirectory {
			fmt.Println("Note: --delete flag not implemented for directory processing yet")
		} else {
			os.Remove(inputFiles[0])
		}
	}
}
