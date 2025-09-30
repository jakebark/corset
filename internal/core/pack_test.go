package core

import (
	"testing"

	"github.com/jakebark/corset/internal/config"
	"github.com/jakebark/corset/internal/inputs"
)

func TestPackAllStatements(t *testing.T) {
	tests := []struct {
		name          string
		userInput     inputs.UserInput
		statements    []Statement
		expectedFiles int
		expectNil     bool
	}{
		{
			name: "Small statements without whitespace",
			userInput: inputs.UserInput{
				Whitespace: false,
				MaxFiles:   5,
			},
			statements: []Statement{
				{Content: map[string]interface{}{"Effect": "Allow"}, Size: 100},
				{Content: map[string]interface{}{"Effect": "Deny"}, Size: 100},
			},
			expectedFiles: 1,
			expectNil:     false,
		},
		{
			name: "Small statements with whitespace",
			userInput: inputs.UserInput{
				Whitespace: true,
				MaxFiles:   5,
			},
			statements: []Statement{
				{Content: map[string]interface{}{"Effect": "Allow"}, Size: 100},
				{Content: map[string]interface{}{"Effect": "Deny"}, Size: 100},
			},
			expectedFiles: 1,
			expectNil:     false,
		},
		{
			name: "Large statements requiring multiple files",
			userInput: inputs.UserInput{
				Whitespace: false,
				MaxFiles:   5,
			},
			statements: []Statement{
				{Content: map[string]interface{}{"Effect": "Allow"}, Size: 3000},
				{Content: map[string]interface{}{"Effect": "Deny"}, Size: 3000},
				{Content: map[string]interface{}{"Effect": "Allow"}, Size: 2000},
			},
			expectedFiles: 2, // Two statements can fit together: 3000+2000=5000 < 5120, one alone: 3000
			expectNil:     false,
		},
		{
			name: "Empty statements",
			userInput: inputs.UserInput{
				Whitespace: false,
				MaxFiles:   5,
			},
			statements:    []Statement{},
			expectedFiles: 0,
			expectNil:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := packAllStatements(tt.userInput, tt.statements)

			if tt.expectNil && result != nil {
				t.Errorf("Expected nil result, got %v", result)
				return
			}

			if !tt.expectNil && result == nil {
				t.Errorf("Expected non-nil result, got nil")
				return
			}

			if len(result) != tt.expectedFiles {
				t.Errorf("Expected %d files, got %d", tt.expectedFiles, len(result))
			}

			// Verify all statements are included
			totalStatements := 0
			for _, file := range result {
				totalStatements += len(file)
			}

			if totalStatements != len(tt.statements) {
				t.Errorf("Expected %d total statements, got %d", len(tt.statements), totalStatements)
			}
		})
	}
}

func TestPackPolicies(t *testing.T) {
	tests := []struct {
		name          string
		statements    []Statement
		baseSize      int
		maxFiles      int
		expectedFiles int
		expectNil     bool
	}{
		{
			name: "Small statements fit in one file",
			statements: []Statement{
				{Content: map[string]interface{}{"Effect": "Allow"}, Size: 100},
				{Content: map[string]interface{}{"Effect": "Deny"}, Size: 100},
			},
			baseSize:      50,
			maxFiles:      5,
			expectedFiles: 1,
			expectNil:     false,
		},
		{
			name: "Large statements require multiple files",
			statements: []Statement{
				{Content: map[string]interface{}{"Effect": "Allow"}, Size: 3000},
				{Content: map[string]interface{}{"Effect": "Deny"}, Size: 3000},
			},
			baseSize:      50,
			maxFiles:      5,
			expectedFiles: 2,
			expectNil:     false,
		},
		{
			name: "Statements too large to fit anywhere",
			statements: []Statement{
				{Content: map[string]interface{}{"Effect": "Allow"}, Size: 6000}, // Exceeds MaxPolicySize
			},
			baseSize:      50,
			maxFiles:      5,
			expectedFiles: 0,
			expectNil:     true,
		},
		{
			name: "Maximum capacity test",
			statements: []Statement{
				{Content: map[string]interface{}{"Effect": "Allow"}, Size: 5000}, // Almost max size
				{Content: map[string]interface{}{"Effect": "Deny"}, Size: 5000},  // Almost max size
				{Content: map[string]interface{}{"Effect": "Allow"}, Size: 100},  // Small
			},
			baseSize:      50,
			maxFiles:      3,
			expectedFiles: 3,
			expectNil:     false,
		},
		{
			name: "Test sorting (largest first)",
			statements: []Statement{
				{Content: map[string]interface{}{"Effect": "Small"}, Size: 100},
				{Content: map[string]interface{}{"Effect": "Large"}, Size: 1000},
				{Content: map[string]interface{}{"Effect": "Medium"}, Size: 500},
			},
			baseSize:      50,
			maxFiles:      5,
			expectedFiles: 1,
			expectNil:     false,
		},
		{
			name:          "Empty statements",
			statements:    []Statement{},
			baseSize:      50,
			maxFiles:      5,
			expectedFiles: 0,
			expectNil:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userInput := inputs.UserInput{
				MaxFiles: tt.maxFiles,
			}

			result := packStatements(userInput, tt.statements, tt.baseSize)

			if tt.expectNil && result != nil {
				t.Errorf("Expected nil result, got %v", result)
				return
			}

			if !tt.expectNil && result == nil {
				t.Errorf("Expected non-nil result, got nil")
				return
			}

			if result != nil {
				if len(result) != tt.expectedFiles {
					t.Errorf("Expected %d files, got %d", tt.expectedFiles, len(result))
				}

				// Verify all statements are included
				totalStatements := 0
				for _, file := range result {
					totalStatements += len(file)
				}

				if totalStatements != len(tt.statements) {
					t.Errorf("Expected %d total statements, got %d", len(tt.statements), totalStatements)
				}

				// Verify no file exceeds the maximum size
				for i, file := range result {
					totalSize := tt.baseSize
					for j, stmt := range file {
						totalSize += stmt.Size
						if j > 0 {
							totalSize += 1 // comma separator
						}
					}

					if totalSize > config.MaxPolicySize {
						t.Errorf("File %d exceeds maximum size: %d > %d", i, totalSize, config.MaxPolicySize)
					}
				}
			}
		})
	}
}

func TestPackPoliciesSorting(t *testing.T) {
	statements := []Statement{
		{Content: map[string]interface{}{"name": "small"}, Size: 100},
		{Content: map[string]interface{}{"name": "large"}, Size: 1000},
		{Content: map[string]interface{}{"name": "medium"}, Size: 500},
		{Content: map[string]interface{}{"name": "tiny"}, Size: 50},
	}

	userInput := inputs.UserInput{
		MaxFiles: 5,
	}

	result := packStatements(userInput, statements, 50)

	if len(result) == 0 {
		t.Fatal("Expected at least one file")
	}

	// First file should contain the largest statement first
	firstFile := result[0]
	if len(firstFile) == 0 {
		t.Fatal("Expected first file to contain statements")
	}

	// The largest statement (size 1000) should be placed first
	if firstFile[0].Size != 1000 {
		t.Errorf("Expected largest statement (1000) to be first, got size %d", firstFile[0].Size)
	}
}

func TestPackPoliciesBinPacking(t *testing.T) {
	// Test the bin packing algorithm effectiveness
	statements := []Statement{
		{Content: map[string]interface{}{"id": "1"}, Size: 2000},
		{Content: map[string]interface{}{"id": "2"}, Size: 2000},
		{Content: map[string]interface{}{"id": "3"}, Size: 1000},
		{Content: map[string]interface{}{"id": "4"}, Size: 1000},
	}

	userInput := inputs.UserInput{
		MaxFiles: 5,
	}

	result := packStatements(userInput, statements, 100)

	if len(result) != 2 {
		t.Errorf("Expected optimal packing into 2 files, got %d", len(result))
	}

	// Verify efficient packing:
	// File 1: 2000 + 1000 + 1000 = 4000 + 100 (base) + 2 (separators) = 4102
	// File 2: 2000 + 100 (base) = 2100
	if len(result) >= 2 {
		file1Size := 100 // base size
		for i, stmt := range result[0] {
			file1Size += stmt.Size
			if i > 0 {
				file1Size += 1 // separator
			}
		}

		file2Size := 100 // base size
		for i, stmt := range result[1] {
			file2Size += stmt.Size
			if i > 0 {
				file2Size += 1 // separator
			}
		}

		// Both files should be under the limit
		if file1Size > config.MaxPolicySize {
			t.Errorf("File 1 exceeds limit: %d > %d", file1Size, config.MaxPolicySize)
		}
		if file2Size > config.MaxPolicySize {
			t.Errorf("File 2 exceeds limit: %d > %d", file2Size, config.MaxPolicySize)
		}
	}
}

