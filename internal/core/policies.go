package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jakebark/corset/internal/config"
	"github.com/jakebark/corset/internal/inputs"
)

type Policy struct {
	Version   string                   `json:"Version"`
	Statement []map[string]interface{} `json:"Statement"`
}

type Statement struct {
	Content map[string]interface{}
	Size    int
}

func ProcessFiles(userInput inputs.UserInput, files []string) {
	allStatements := extractAllStatements(files)
	if len(allStatements) == 0 {
		fmt.Println("No policy statements found")
		return
	}

	packedFiles := packAllStatements(userInput, allStatements)
	writeOutputFiles(userInput, packedFiles, files)
}

func extractAllStatements(files []string) []Statement {
	var allStatements []Statement
	for _, file := range files {
		statements := extractIndividualPolicies(file)
		allStatements = append(allStatements, statements...)
	}
	return allStatements
}

func extractIndividualPolicies(filename string) []Statement {
	data, _ := os.ReadFile(filename)

	var policy Policy
	json.Unmarshal(data, &policy)

	var statements []Statement
	for _, stmt := range policy.Statement {
		stmtJSON, _ := json.Marshal(stmt)

		statements = append(statements, Statement{
			Content: stmt,
			Size:    len(stmtJSON),
		})
	}

	return statements
}

func packAllStatements(userInput inputs.UserInput, statements []Statement) [][]Statement {
	baseSize := config.SCPBaseSizeMinified
	if userInput.Whitespace {
		baseSize = config.SCPBaseSizeWithWS
	}
	return packPolicies(userInput, statements, baseSize)
}

// first fit / bin pack
func packPolicies(userInput inputs.UserInput, statements []Statement, baseSize int) [][]Statement {
	// Sort policies by size (largest first) for better bin packing
	sort.Slice(statements, func(i, j int) bool {
		return statements[i].Size > statements[j].Size
	})

	files := make([][]Statement, userInput.MaxFiles)
	fileSizes := make([]int, userInput.MaxFiles)

	// Initialize each file with base structure size
	for i := range fileSizes {
		fileSizes[i] = baseSize
	}

	// First Fit Decreasing algorithm
	for _, stmt := range statements {
		placed := false

		// Try to place in existing file with space
		for i := 0; i < userInput.MaxFiles; i++ {
			// Account for comma separator (except for first statement)
			separator := 0
			if len(files[i]) > 0 {
				separator = 1 // for comma
			}

			if fileSizes[i]+stmt.Size+separator <= config.MaxPolicySize {
				files[i] = append(files[i], stmt)
				fileSizes[i] += stmt.Size + separator
				placed = true
				break
			}
		}

		if !placed {
			return nil // Cannot fit all policies
		}
	}

	// Remove empty files
	var result [][]Statement
	for _, file := range files {
		if len(file) > 0 {
			result = append(result, file)
		}
	}

	return result
}


func writeOutputFiles(userInput inputs.UserInput, packedFiles [][]Statement, inputFiles []string) {
	outputDir := filepath.Dir(inputFiles[0])
	
	fmt.Printf("Split into %d files:\n", len(packedFiles))
	for i, statements := range packedFiles {
		filename := filepath.Join(outputDir, fmt.Sprintf("corset%d.json", i+1))
		size := writeOutputFile(userInput, filename, statements)
		fmt.Printf("- %s (%d characters, %d statements)\n", filepath.Base(filename), size, len(statements))
	}
	
	if userInput.Delete {
		if userInput.IsDirectory {
			fmt.Println("Note: --delete flag not implemented for directory processing yet")
		} else {
			os.Remove(inputFiles[0])
		}
	}
}


func writeOutputFile(userInput inputs.UserInput, filename string, statements []Statement) int {
	policy := Policy{
		Version:   "2012-10-17",
		Statement: make([]map[string]interface{}, len(statements)),
	}

	for i, stmt := range statements {
		policy.Statement[i] = stmt.Content
	}

	var data []byte

	if userInput.Whitespace {
		data, _ = json.MarshalIndent(policy, "", "  ")
	} else {
		data, _ = json.Marshal(policy)
	}

	os.WriteFile(filename, data, 0644)

	return len(data)
}
