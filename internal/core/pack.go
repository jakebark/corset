package core

import (
	"sort"

	"github.com/jakebark/corset/internal/config"
	"github.com/jakebark/corset/internal/inputs"
)

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

	// Return empty slice instead of nil for empty results
	if result == nil {
		result = [][]Statement{}
	}

	return result
}
