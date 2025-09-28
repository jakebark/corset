package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jakebark/corset/internal/config"
	"github.com/jakebark/corset/internal/inputs"
)

func TestGenerateOutputFilename(t *testing.T) {
	tests := []struct {
		name       string
		userInput  inputs.UserInput
		outputDir  string
		fileNum    int
		inputFiles []string
		expected   string
	}{
		{
			name: "Default naming convention",
			userInput: inputs.UserInput{
				Replace:     false,
				IsDirectory: false,
				Target:      "/path/to/file.json",
			},
			outputDir:  "/output",
			fileNum:    1,
			inputFiles: []string{"/path/to/file.json"},
			expected:   "/output/corset1.json",
		},
		{
			name: "Default naming convention - multiple files",
			userInput: inputs.UserInput{
				Replace:     false,
				IsDirectory: false,
				Target:      "/path/to/file.json",
			},
			outputDir:  "/output",
			fileNum:    3,
			inputFiles: []string{"/path/to/file.json"},
			expected:   "/output/corset3.json",
		},
		{
			name: "Single file replacement",
			userInput: inputs.UserInput{
				Replace:     true,
				IsDirectory: false,
				Target:      "/path/to/policy.json",
			},
			outputDir:  "/output",
			fileNum:    1,
			inputFiles: []string{"/path/to/policy.json"},
			expected:   "/path/to/policy.json",
		},
		{
			name: "Directory replacement - first file",
			userInput: inputs.UserInput{
				Replace:     true,
				IsDirectory: true,
				Target:      "/path/to/organisation-scp",
			},
			outputDir:  "/path/to/organisation-scp",
			fileNum:    1,
			inputFiles: []string{"/path/to/organisation-scp/policy1.json", "/path/to/organisation-scp/policy2.json"},
			expected:   "/path/to/organisation-scp/organisation-scp.json",
		},
		{
			name: "Directory replacement - second file",
			userInput: inputs.UserInput{
				Replace:     true,
				IsDirectory: true,
				Target:      "/path/to/organisation-scp",
			},
			outputDir:  "/path/to/organisation-scp",
			fileNum:    2,
			inputFiles: []string{"/path/to/organisation-scp/policy1.json", "/path/to/organisation-scp/policy2.json"},
			expected:   "/path/to/organisation-scp/organisation-scp-2.json",
		},
		{
			name: "Directory replacement - third file",
			userInput: inputs.UserInput{
				Replace:     true,
				IsDirectory: true,
				Target:      "/path/to/my-policies",
			},
			outputDir:  "/path/to/my-policies",
			fileNum:    3,
			inputFiles: []string{"/path/to/my-policies/a.json", "/path/to/my-policies/b.json"},
			expected:   "/path/to/my-policies/my-policies-3.json",
		},
		{
			name: "Replace with multiple single files (fallback to default)",
			userInput: inputs.UserInput{
				Replace:     true,
				IsDirectory: false,
				Target:      "/path/to/policy.json",
			},
			outputDir:  "/output",
			fileNum:    1,
			inputFiles: []string{"/path/to/policy1.json", "/path/to/policy2.json"}, // Multiple files but not directory
			expected:   "/output/corset1.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateOutputFilename(tt.userInput, tt.outputDir, tt.fileNum, tt.inputFiles)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSingleFileReplacement(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create test input file
	inputFile := filepath.Join(tempDir, "policy.json")
	policy := testPolicy{
		Version: "2012-10-17",
		Statement: []map[string]interface{}{
			{"Effect": "Allow", "Action": "s3:GetObject", "Resource": "*"},
			{"Effect": "Deny", "Action": "s3:DeleteObject", "Resource": "*"},
		},
	}
	
	data, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test policy: %v", err)
	}
	
	err = os.WriteFile(inputFile, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	// Store original size
	originalInfo, err := os.Stat(inputFile)
	if err != nil {
		t.Fatalf("Failed to stat original file: %v", err)
	}
	originalSize := originalInfo.Size()
	
	userInput := inputs.UserInput{
		Target:      inputFile,
		Replace:     true,
		Whitespace:  false, // Should minify
		IsDirectory: false,
		MaxFiles:    config.DefaultMaxFiles,
	}
	
	ProcessFiles(userInput, []string{inputFile})
	
	// Verify original file still exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		t.Fatal("Expected original file to still exist after replacement")
	}
	
	// Verify no corset files were created
	corsetFile := filepath.Join(tempDir, "corset1.json")
	if _, err := os.Stat(corsetFile); err == nil {
		t.Error("Expected no corset1.json file for single file replacement")
	}
	
	// Verify file content is valid and minified
	newData, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("Failed to read replaced file: %v", err)
	}
	
	var newPolicy testPolicy
	err = json.Unmarshal(newData, &newPolicy)
	if err != nil {
		t.Fatalf("Replaced file contains invalid JSON: %v", err)
	}
	
	// Verify content matches
	if newPolicy.Version != policy.Version {
		t.Errorf("Version mismatch: expected %s, got %s", policy.Version, newPolicy.Version)
	}
	
	if len(newPolicy.Statement) != len(policy.Statement) {
		t.Errorf("Statement count mismatch: expected %d, got %d", len(policy.Statement), len(newPolicy.Statement))
	}
	
	// Verify file was minified (should be smaller)
	newInfo, err := os.Stat(inputFile)
	if err != nil {
		t.Fatalf("Failed to stat replaced file: %v", err)
	}
	newSize := newInfo.Size()
	
	if newSize >= originalSize {
		t.Errorf("Expected minified file to be smaller: original %d, new %d", originalSize, newSize)
	}
}

func TestDirectoryReplacement(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create a subdirectory with a specific name
	targetDir := filepath.Join(tempDir, "organisation-scp")
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}
	
	// Create multiple test files
	policies := []testPolicy{
		{
			Version: "2012-10-17",
			Statement: []map[string]interface{}{
				{"Effect": "Allow", "Action": "s3:GetObject", "Resource": "*"},
			},
		},
		{
			Version: "2012-10-17",
			Statement: []map[string]interface{}{
				{"Effect": "Deny", "Action": "s3:DeleteObject", "Resource": "*"},
			},
		},
	}
	
	var inputFiles []string
	for i, policy := range policies {
		filename := filepath.Join(targetDir, filepath.Base(tempDir)+"-policy-"+string(rune('a'+i))+".json")
		data, err := json.MarshalIndent(policy, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal test policy %d: %v", i, err)
		}
		
		err = os.WriteFile(filename, data, 0644)
		if err != nil {
			t.Fatalf("Failed to write test file %d: %v", i, err)
		}
		inputFiles = append(inputFiles, filename)
	}
	
	userInput := inputs.UserInput{
		Target:      targetDir,
		Replace:     true,
		Whitespace:  false,
		IsDirectory: true,
		MaxFiles:    config.DefaultMaxFiles,
	}
	
	ProcessFiles(userInput, inputFiles)
	
	// Verify original files were deleted
	for _, file := range inputFiles {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			t.Errorf("Expected original file %s to be deleted", file)
		}
	}
	
	// Verify new files with correct naming
	expectedFile := filepath.Join(targetDir, "organisation-scp.json")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected output file %s to be created", expectedFile)
	}
	
	// Verify no corset files were created
	corsetFile := filepath.Join(targetDir, "corset1.json")
	if _, err := os.Stat(corsetFile); err == nil {
		t.Error("Expected no corset1.json file for directory replacement")
	}
	
	// Verify content is valid
	data, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	
	var combinedPolicy testPolicy
	err = json.Unmarshal(data, &combinedPolicy)
	if err != nil {
		t.Fatalf("Output file contains invalid JSON: %v", err)
	}
	
	// Should have combined statements from both input files
	if len(combinedPolicy.Statement) != 2 {
		t.Errorf("Expected 2 combined statements, got %d", len(combinedPolicy.Statement))
	}
}

func TestDirectoryReplacementMultipleFiles(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create a subdirectory
	targetDir := filepath.Join(tempDir, "large-policies")
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}
	
	// Create a policy large enough to require splitting
	largePolicy := testPolicy{
		Version:   "2012-10-17",
		Statement: createLargeStatements(15), // Should require splitting
	}
	
	inputFile := filepath.Join(targetDir, "large.json")
	data, err := json.MarshalIndent(largePolicy, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal large policy: %v", err)
	}
	
	err = os.WriteFile(inputFile, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write large policy file: %v", err)
	}
	
	userInput := inputs.UserInput{
		Target:      targetDir,
		Replace:     true,
		Whitespace:  false,
		IsDirectory: true,
		MaxFiles:    config.DefaultMaxFiles,
	}
	
	ProcessFiles(userInput, []string{inputFile})
	
	// Verify original file was deleted
	if _, err := os.Stat(inputFile); !os.IsNotExist(err) {
		t.Error("Expected original large file to be deleted")
	}
	
	// Verify multiple output files with correct naming
	expectedFiles := []string{
		filepath.Join(targetDir, "large-policies.json"),     // First file
		filepath.Join(targetDir, "large-policies-2.json"),  // Second file
	}
	
	foundFiles := 0
	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(expectedFile); err == nil {
			foundFiles++
			
			// Verify content is valid
			data, err := os.ReadFile(expectedFile)
			if err != nil {
				t.Errorf("Failed to read output file %s: %v", expectedFile, err)
				continue
			}
			
			var policy testPolicy
			err = json.Unmarshal(data, &policy)
			if err != nil {
				t.Errorf("Output file %s contains invalid JSON: %v", expectedFile, err)
			}
		}
	}
	
	if foundFiles < 2 {
		t.Errorf("Expected at least 2 output files, found %d", foundFiles)
	}
	
	// Verify no corset files were created
	corsetFile := filepath.Join(targetDir, "corset1.json")
	if _, err := os.Stat(corsetFile); err == nil {
		t.Error("Expected no corset1.json file for directory replacement")
	}
}