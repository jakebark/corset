package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jakebark/corset/internal/config"
	"github.com/jakebark/corset/internal/core"
	"github.com/jakebark/corset/internal/inputs"
)

type Policy struct {
	Version   string                   `json:"Version"`
	Statement []map[string]interface{} `json:"Statement"`
}

type integrationTestCase struct {
	name            string
	whitespace      bool
	isDirectory     bool
	inputFiles      []string // testdata files to use
	expectedFiles   int      // expected number of output files
	expectSplit     bool     // whether we expect policy to be split
	verifyIntegrity bool     // whether to verify no statements are cut
}

// TestEndToEndPolicyProcessing tests the complete workflow from CLI input to file output
// This is the main integration test that verifies the entire corset pipeline works correctly
func TestEndToEndPolicyProcessing(t *testing.T) {
	tests := []integrationTestCase{
		{
			name:            "Single small policy no whitespace",
			whitespace:      false,
			isDirectory:     false,
			inputFiles:      []string{"small_policy.json"},
			expectedFiles:   1,
			expectSplit:     false,
			verifyIntegrity: true,
		},
		{
			name:            "Single small policy with whitespace",
			whitespace:      true,
			isDirectory:     false,
			inputFiles:      []string{"small_policy.json"},
			expectedFiles:   1,
			expectSplit:     false,
			verifyIntegrity: true,
		},
		{
			name:            "Large policy forces split",
			whitespace:      false,
			isDirectory:     false,
			inputFiles:      []string{"very_large_policy.json"},
			expectedFiles:   2, // Should split into multiple files
			expectSplit:     true,
			verifyIntegrity: true,
		},
		{
			name:            "Directory with multiple files",
			whitespace:      false,
			isDirectory:     true,
			inputFiles:      []string{"policy1.json", "policy2.json", "policy3.json"},
			expectedFiles:   1, // Should combine into one file
			expectSplit:     false,
			verifyIntegrity: true,
		},
		{
			name:            "Policy integrity test - statements not cut",
			whitespace:      false,
			isDirectory:     false,
			inputFiles:      []string{"integrity_test.json"},
			expectedFiles:   1, // Should fit in one file
			expectSplit:     false,
			verifyIntegrity: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Setup input files from testdata
			var targetPath string
			if tt.isDirectory {
				targetPath = tempDir
				for _, filename := range tt.inputFiles {
					copyTestDataFile(t, filename, tempDir)
				}
			} else {
				// Single file
				filename := tt.inputFiles[0]
				targetPath = copyTestDataFile(t, filename, tempDir)
			}

			// Create UserInput
			userInput := inputs.UserInput{
				Target:      targetPath,
				Whitespace:  tt.whitespace,
				IsDirectory: tt.isDirectory,
				MaxFiles:    config.DefaultMaxFiles,
			}

			// Process files
			var files []string
			if tt.isDirectory {
				files = core.FindJSONFilesInDirectory(targetPath)
			} else {
				files = []string{targetPath}
			}

			core.ProcessFiles(userInput, files)

			// Verify results
			outputFiles := findOutputFiles(tempDir, tt.isDirectory, filepath.Base(targetPath))

			// Check expected number of files
			if len(outputFiles) != tt.expectedFiles {
				t.Errorf("Expected %d output files, got %d", tt.expectedFiles, len(outputFiles))
			}

			// Check split behavior
			if tt.expectSplit && len(outputFiles) <= 1 {
				t.Error("Expected policy to be split across multiple files")
			}
			if !tt.expectSplit && len(outputFiles) > 1 {
				t.Error("Expected policy to fit in single file")
			}

			// Check whitespace formatting
			if len(outputFiles) > 0 {
				content := readFileContent(t, outputFiles[0])
				hasWhitespace := containsWhitespace(content)
				if tt.whitespace && !hasWhitespace {
					t.Error("Expected whitespace formatting in output")
				}
				if !tt.whitespace && hasWhitespace {
					t.Error("Expected no whitespace formatting in output")
				}
			}

			// Verify policy integrity
			if tt.verifyIntegrity {
				inputPolicies := loadInputPolicies(t, tt.inputFiles)
				verifyPolicyIntegrity(t, inputPolicies, outputFiles)
			}

			// Cleanup
			for _, file := range outputFiles {
				os.Remove(file)
			}
		})
	}
}

// Helper functions

func copyTestDataFile(t *testing.T, filename, destDir string) string {
	testDataPath := filepath.Join("testdata", filename)
	data, err := os.ReadFile(testDataPath)
	if err != nil {
		t.Fatalf("Failed to read testdata file %s: %v", testDataPath, err)
	}

	destPath := filepath.Join(destDir, filename)
	err = os.WriteFile(destPath, data, 0644)
	if err != nil {
		t.Fatalf("Failed to copy testdata file to %s: %v", destPath, err)
	}

	return destPath
}

func loadInputPolicies(t *testing.T, filenames []string) []Policy {
	var policies []Policy
	for _, filename := range filenames {
		testDataPath := filepath.Join("testdata", filename)
		data, err := os.ReadFile(testDataPath)
		if err != nil {
			t.Fatalf("Failed to read testdata file %s: %v", testDataPath, err)
		}

		var policy Policy
		err = json.Unmarshal(data, &policy)
		if err != nil {
			t.Fatalf("Failed to unmarshal testdata file %s: %v", testDataPath, err)
		}

		policies = append(policies, policy)
	}
	return policies
}

func findOutputFiles(tempDir string, isDirectory bool, baseName string) []string {
	var outputFiles []string

	if isDirectory {
		// Directory replacement: look for baseName.json, baseName-2.json, etc.
		dirName := filepath.Base(tempDir)
		for i := 1; i <= 5; i++ {
			var outputFile string
			if i == 1 {
				outputFile = filepath.Join(tempDir, dirName+".json")
			} else {
				outputFile = filepath.Join(tempDir, fmt.Sprintf("%s-%d.json", dirName, i))
			}
			if _, err := os.Stat(outputFile); err == nil {
				outputFiles = append(outputFiles, outputFile)
			}
		}
	} else {
		// Single file replacement: look for original filename (replaced in-place)
		// Find any .json files in the directory (should be the replaced file)
		files, _ := filepath.Glob(filepath.Join(tempDir, "*.json"))
		outputFiles = append(outputFiles, files...)
	}

	return outputFiles
}

func readFileContent(t *testing.T, filepath string) string {
	data, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filepath, err)
	}
	return string(data)
}

func readOutputFile(t *testing.T, filepath string) Policy {
	data, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Failed to read output file %s: %v", filepath, err)
	}

	var policy Policy
	err = json.Unmarshal(data, &policy)
	if err != nil {
		t.Fatalf("Failed to unmarshal output file %s: %v", filepath, err)
	}

	return policy
}

func containsWhitespace(content string) bool {
	// Check for indentation and newlines that indicate formatting
	return strings.Contains(content, "\n  ") || strings.Contains(content, "{\n  ")
}

func verifyPolicyIntegrity(t *testing.T, inputPolicies []Policy, outputFiles []string) {
	// Collect all input statements
	expectedStatements := make(map[string]map[string]interface{})
	for _, policy := range inputPolicies {
		for _, stmt := range policy.Statement {
			// Use a key based on statement content for uniqueness
			key := fmt.Sprintf("%v", stmt)
			expectedStatements[key] = stmt
		}
	}

	// Collect all output statements
	foundStatements := make(map[string]map[string]interface{})
	for _, outputFile := range outputFiles {
		policy := readOutputFile(t, outputFile)

		if policy.Version != "2012-10-17" {
			t.Errorf("Invalid version in output file %s", outputFile)
		}

		for _, stmt := range policy.Statement {
			key := fmt.Sprintf("%v", stmt)
			if _, exists := foundStatements[key]; exists {
				t.Errorf("Duplicate statement found in output: %v", stmt)
			}
			foundStatements[key] = stmt
		}
	}

	// Verify all statements are preserved
	if len(expectedStatements) != len(foundStatements) {
		t.Errorf("Statement count mismatch: expected %d, got %d",
			len(expectedStatements), len(foundStatements))
	}

	for key := range expectedStatements {
		if _, found := foundStatements[key]; !found {
			t.Errorf("Statement missing in output: %v", expectedStatements[key])
		}
	}
}
