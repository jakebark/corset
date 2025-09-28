package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jakebark/corset/internal/config"
	"github.com/jakebark/corset/internal/inputs"
)

func TestProcessFiles(t *testing.T) {
	tests := []struct {
		name      string
		userInput inputs.UserInput
		policies  []testPolicy
		expectOutput bool
	}{
		{
			name: "Single file with statements",
			userInput: inputs.UserInput{
				Replace:      false,
				Whitespace:  false,
				IsDirectory: false,
				MaxFiles:    config.DefaultMaxFiles,
			},
			policies: []testPolicy{
				{
					Version: "2012-10-17",
					Statement: []map[string]interface{}{
						{"Effect": "Allow", "Action": "s3:GetObject", "Resource": "*"},
						{"Effect": "Deny", "Action": "s3:DeleteObject", "Resource": "*"},
					},
				},
			},
			expectOutput: true,
		},
		{
			name: "Multiple files with statements",
			userInput: inputs.UserInput{
				Replace:      false,
				Whitespace:  true,
				IsDirectory: true,
				MaxFiles:    config.DefaultMaxFiles,
			},
			policies: []testPolicy{
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
			},
			expectOutput: true,
		},
		{
			name: "Empty policy files",
			userInput: inputs.UserInput{
				Replace:      false,
				Whitespace:  false,
				IsDirectory: false,
				MaxFiles:    config.DefaultMaxFiles,
			},
			policies: []testPolicy{
				{
					Version:   "2012-10-17",
					Statement: []map[string]interface{}{},
				},
			},
			expectOutput: false,
		},
		{
			name: "Large policy requiring splitting",
			userInput: inputs.UserInput{
				Replace:      false,
				Whitespace:  false,
				IsDirectory: false,
				MaxFiles:    config.DefaultMaxFiles,
			},
			policies: []testPolicy{
				{
					Version: "2012-10-17",
					Statement: createLargeStatements(20), // Large enough to require splitting
				},
			},
			expectOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			var files []string
			
			// Create test files
			for i, policy := range tt.policies {
				filename := filepath.Join(tempDir, "policy_"+string(rune('a'+i))+".json")
				data, err := json.MarshalIndent(policy, "", "  ")
				if err != nil {
					t.Fatalf("Failed to marshal test policy %d: %v", i, err)
				}
				
				err = os.WriteFile(filename, data, 0644)
				if err != nil {
					t.Fatalf("Failed to write test file %d: %v", i, err)
				}
				files = append(files, filename)
			}
			
			// Set target for userInput
			if len(files) > 0 {
				if tt.userInput.IsDirectory {
					tt.userInput.Target = tempDir
				} else {
					tt.userInput.Target = files[0]
				}
			}
			
			// Test the function - should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("ProcessFiles panicked: %v", r)
				}
			}()
			
			ProcessFiles(tt.userInput, files)
			
			if tt.expectOutput {
				// Check that output files were created
				foundOutput := false
				for i := 1; i <= 5; i++ { // Check for corset1.json, corset2.json, etc.
					outputFile := filepath.Join(tempDir, "corset"+string(rune('0'+i))+".json")
					if _, err := os.Stat(outputFile); err == nil {
						foundOutput = true
						
						// Verify the output file is valid JSON
						data, err := os.ReadFile(outputFile)
						if err != nil {
							t.Errorf("Failed to read output file %s: %v", outputFile, err)
							continue
						}
						
						var policy testPolicy
						err = json.Unmarshal(data, &policy)
						if err != nil {
							t.Errorf("Output file %s contains invalid JSON: %v", outputFile, err)
							continue
						}
						
						// Verify structure
						if policy.Version != config.SCPVersion {
							t.Errorf("Output file %s has wrong version: expected %s, got %s", 
								outputFile, config.SCPVersion, policy.Version)
						}
					}
				}
				
				if !foundOutput {
					t.Error("Expected output files to be created, but none were found")
				}
			}
		})
	}
}

func TestProcessFilesErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		userInput inputs.UserInput
		files     []string
		setupFunc func(t *testing.T, tempDir string) []string
	}{
		{
			name: "Non-existent files",
			userInput: inputs.UserInput{
				Replace:      false,
				Whitespace:  false,
				IsDirectory: false,
				MaxFiles:    config.DefaultMaxFiles,
			},
			setupFunc: func(t *testing.T, tempDir string) []string {
				return []string{filepath.Join(tempDir, "nonexistent.json")}
			},
		},
		{
			name: "Invalid JSON files",
			userInput: inputs.UserInput{
				Replace:      false,
				Whitespace:  false,
				IsDirectory: false,
				MaxFiles:    config.DefaultMaxFiles,
			},
			setupFunc: func(t *testing.T, tempDir string) []string {
				filename := filepath.Join(tempDir, "invalid.json")
				err := os.WriteFile(filename, []byte(`{"invalid": json}`), 0644)
				if err != nil {
					t.Fatalf("Failed to write invalid JSON file: %v", err)
				}
				return []string{filename}
			},
		},
		{
			name: "Empty file list",
			userInput: inputs.UserInput{
				Replace:      false,
				Whitespace:  false,
				IsDirectory: false,
				MaxFiles:    config.DefaultMaxFiles,
			},
			setupFunc: func(t *testing.T, tempDir string) []string {
				return []string{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			files := tt.setupFunc(t, tempDir)
			
			// Should not panic, should handle gracefully
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("ProcessFiles panicked on error case: %v", r)
				}
			}()
			
			ProcessFiles(tt.userInput, files)
		})
	}
}

func TestProcessFilesWithReplacement(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create test file
	testFile := filepath.Join(tempDir, "test.json")
	policy := testPolicy{
		Version: "2012-10-17",
		Statement: []map[string]interface{}{
			{"Effect": "Allow", "Action": "s3:GetObject", "Resource": "*"},
		},
	}
	
	data, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test policy: %v", err)
	}
	
	err = os.WriteFile(testFile, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	userInput := inputs.UserInput{
		Target:      testFile,
		Replace:     true,
		Whitespace:  false,
		IsDirectory: false,
		MaxFiles:    config.DefaultMaxFiles,
	}
	
	ProcessFiles(userInput, []string{testFile})
	
	// Verify original file was replaced (should still exist with new content)
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Expected original file to be replaced, but it doesn't exist")
	}
	
	// For single file replacement, no separate corset1.json should be created
	outputFile := filepath.Join(tempDir, "corset1.json")
	if _, err := os.Stat(outputFile); err == nil {
		t.Error("Expected no separate corset1.json file for single file replacement")
	}
	
	// Verify the content of the replaced file is valid JSON
	newData, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read replaced file: %v", err)
	}
	
	var newPolicy testPolicy
	err = json.Unmarshal(newData, &newPolicy)
	if err != nil {
		t.Fatalf("Replaced file contains invalid JSON: %v", err)
	}
}

// Helper function to create large statements for testing
func createLargeStatements(count int) []map[string]interface{} {
	statements := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		statements[i] = map[string]interface{}{
			"Sid":    "Statement" + string(rune('0'+i)),
			"Effect": "Allow",
			"Action": []string{
				"s3:GetObject", "s3:PutObject", "s3:DeleteObject", 
				"s3:ListBucket", "s3:GetBucketLocation",
			},
			"Resource": []string{
				"arn:aws:s3:::very-long-bucket-name-that-takes-up-significant-space-" + string(rune('0'+i)) + "/*",
				"arn:aws:s3:::very-long-bucket-name-that-takes-up-significant-space-" + string(rune('0'+i)),
			},
			"Condition": map[string]interface{}{
				"StringEquals": map[string]interface{}{
					"aws:userid": "AIDACKCEVSQ6C2EXAMPLE:very-long-user-identifier-that-consumes-lots-of-characters-" + string(rune('0'+i)),
				},
			},
		}
	}
	return statements
}