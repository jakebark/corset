package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jakebark/corset/internal/config"
	"github.com/jakebark/corset/internal/inputs"
)

type testPolicy struct {
	Version   string                   `json:"Version"`
	Statement []map[string]interface{} `json:"Statement"`
}

func TestProcessor_extractIndividualPolicies(t *testing.T) {
	tests := []struct {
		name               string
		policy             testPolicy
		expectedStatements int
	}{
		{
			name: "Small policy with 2 statements",
			policy: testPolicy{
				Version: "2012-10-17",
				Statement: []map[string]interface{}{
					{"Effect": "Allow", "Action": "s3:GetObject", "Resource": "*"},
					{"Effect": "Deny", "Action": "s3:DeleteObject", "Resource": "*"},
				},
			},
			expectedStatements: 2,
		},
		{
			name: "Empty policy",
			policy: testPolicy{
				Version:   "2012-10-17",
				Statement: []map[string]interface{}{},
			},
			expectedStatements: 0,
		},
		{
			name: "Single statement policy",
			policy: testPolicy{
				Version: "2012-10-17",
				Statement: []map[string]interface{}{
					{"Effect": "Allow", "Action": "s3:*", "Resource": "*"},
				},
			},
			expectedStatements: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			
			// Create test file
			testFile := filepath.Join(tempDir, "test.json")
			data, err := json.MarshalIndent(tt.policy, "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal test policy: %v", err)
			}
			
			err = os.WriteFile(testFile, data, 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}
			
			// Test the function
			userInput := inputs.UserInput{
				Target:      testFile,
				Delete:      false,
				Whitespace:  false,
				IsDirectory: false,
				MaxFiles:    config.DefaultMaxFiles,
			}
			
			processor := NewProcessor(userInput)
			statements := processor.extractIndividualPolicies(testFile)
			
			if len(statements) != tt.expectedStatements {
				t.Errorf("Expected %d statements, got %d", tt.expectedStatements, len(statements))
			}
			
			// Verify each statement has required fields
			for i, stmt := range statements {
				if stmt.Content == nil {
					t.Errorf("Statement %d has nil content", i)
				}
				if stmt.Size <= 0 {
					t.Errorf("Statement %d has invalid size: %d", i, stmt.Size)
				}
				// OriginalFilename field removed for simplification
			}
		})
	}
}

func TestBaseSizeConstants(t *testing.T) {
	tests := []struct {
		name       string
		whitespace bool
		expected   int
	}{
		{
			name:       "With whitespace",
			whitespace: true,
			expected:   config.SCPBaseSizeWithWS,
		},
		{
			name:       "Without whitespace",
			whitespace: false,
			expected:   config.SCPBaseSizeMinified,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result int
			if tt.whitespace {
				result = config.SCPBaseSizeWithWS
			} else {
				result = config.SCPBaseSizeMinified
			}
			
			if result != tt.expected {
				t.Errorf("Expected base size %d, got %d", tt.expected, result)
			}
			
			// Verify constants match actual string lengths
			if tt.whitespace {
				actualSize := len(config.SCPBaseWithWS) - 2
				if result != actualSize {
					t.Errorf("Constant SCPBaseSizeWithWS (%d) doesn't match actual size (%d)", result, actualSize)
				}
			} else {
				actualSize := len(config.SCPBaseStructure) - 2
				if result != actualSize {
					t.Errorf("Constant SCPBaseSizeMinified (%d) doesn't match actual size (%d)", result, actualSize)
				}
			}
		})
	}
}

func TestProcessor_writeOutputFile(t *testing.T) {
	tests := []struct {
		name       string
		whitespace bool
		statements []Statement
	}{
		{
			name:       "Single statement with whitespace",
			whitespace: true,
			statements: []Statement{
				{
					Content: map[string]interface{}{
						"Effect":   "Allow",
						"Action":   "s3:GetObject",
						"Resource": "*",
					},
					Size: 50,
				},
			},
		},
		{
			name:       "Multiple statements without whitespace",
			whitespace: false,
			statements: []Statement{
				{
					Content: map[string]interface{}{
						"Effect":   "Allow",
						"Action":   "s3:GetObject",
						"Resource": "*",
					},
					Size: 50,
				},
				{
					Content: map[string]interface{}{
						"Effect":   "Deny",
						"Action":   "s3:DeleteObject",
						"Resource": "*",
					},
					Size: 50,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			outputFile := filepath.Join(tempDir, "output.json")
			
			userInput := inputs.UserInput{
				Target:      "/test",
				Delete:      false,
				Whitespace:  tt.whitespace,
				IsDirectory: false,
				MaxFiles:    config.DefaultMaxFiles,
			}
			
			processor := NewProcessor(userInput)
			size := processor.writeOutputFile(outputFile, tt.statements)
			
			// Verify file was created
			if _, err := os.Stat(outputFile); os.IsNotExist(err) {
				t.Fatal("Output file was not created")
			}
			
			// Verify file content
			data, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}
			
			content := string(data)
			
			// Verify size matches
			if size != len(data) {
				t.Errorf("Expected size %d, got %d", len(data), size)
			}
			
			// Verify whitespace formatting
			hasWhitespace := strings.Contains(content, "\n  ")
			if tt.whitespace && !hasWhitespace {
				t.Error("Expected whitespace formatting")
			}
			if !tt.whitespace && hasWhitespace {
				t.Error("Expected no whitespace formatting")
			}
			
			// Verify valid JSON
			var policy testPolicy
			err = json.Unmarshal(data, &policy)
			if err != nil {
				t.Fatalf("Output is not valid JSON: %v", err)
			}
			
			// Verify structure
			if policy.Version != "2012-10-17" {
				t.Errorf("Expected version 2012-10-17, got %s", policy.Version)
			}
			
			if len(policy.Statement) != len(tt.statements) {
				t.Errorf("Expected %d statements, got %d", len(tt.statements), len(policy.Statement))
			}
		})
	}
}

func TestProcessor_packPolicies(t *testing.T) {
	tests := []struct {
		name          string
		statements    []Statement
		baseSize      int
		maxFiles      int
		expectedFiles int
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userInput := inputs.UserInput{
				Target:      "/test",
				Delete:      false,
				Whitespace:  false,
				IsDirectory: false,
				MaxFiles:    tt.maxFiles,
			}
			
			processor := NewProcessor(userInput)
			result := processor.packPolicies(tt.statements, tt.baseSize)
			
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