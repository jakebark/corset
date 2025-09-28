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

func TestCreatePolicyJSON(t *testing.T) {
	tests := []struct {
		name       string
		userInput  inputs.UserInput
		statements []Statement
		wantIndent bool
	}{
		{
			name: "Single statement without whitespace",
			userInput: inputs.UserInput{
				Whitespace: false,
			},
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
			wantIndent: false,
		},
		{
			name: "Single statement with whitespace",
			userInput: inputs.UserInput{
				Whitespace: true,
			},
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
			wantIndent: true,
		},
		{
			name: "Multiple statements",
			userInput: inputs.UserInput{
				Whitespace: false,
			},
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
			wantIndent: false,
		},
		{
			name: "Empty statements",
			userInput: inputs.UserInput{
				Whitespace: false,
			},
			statements: []Statement{},
			wantIndent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := createPolicyJSON(tt.userInput, tt.statements)
			
			// Verify it's valid JSON
			var policy testPolicy
			err := json.Unmarshal(data, &policy)
			if err != nil {
				t.Fatalf("Generated invalid JSON: %v", err)
			}
			
			// Verify structure
			if policy.Version != config.SCPVersion {
				t.Errorf("Expected version %s, got %s", config.SCPVersion, policy.Version)
			}
			
			if len(policy.Statement) != len(tt.statements) {
				t.Errorf("Expected %d statements, got %d", len(tt.statements), len(policy.Statement))
			}
			
			// Verify whitespace formatting
			content := string(data)
			hasIndent := strings.Contains(content, "\n  ")
			if tt.wantIndent && !hasIndent {
				t.Error("Expected indented formatting")
			}
			if !tt.wantIndent && hasIndent {
				t.Error("Expected minified formatting")
			}
			
			// Verify statement content
			for i, stmt := range tt.statements {
				if i < len(policy.Statement) {
					originalContent := stmt.Content
					generatedContent := policy.Statement[i]
					
					if !mapsEqual(originalContent, generatedContent) {
						t.Errorf("Statement %d content mismatch", i)
					}
				}
			}
		})
	}
}

func TestWriteOutputFile(t *testing.T) {
	tests := []struct {
		name       string
		userInput  inputs.UserInput
		statements []Statement
		filename   string
	}{
		{
			name: "Single statement with whitespace",
			userInput: inputs.UserInput{
				Whitespace: true,
			},
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
			filename: "output_ws.json",
		},
		{
			name: "Multiple statements without whitespace",
			userInput: inputs.UserInput{
				Whitespace: false,
			},
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
			filename: "output_min.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			outputFile := filepath.Join(tempDir, tt.filename)
			
			size := writeOutputFile(tt.userInput, outputFile, tt.statements)
			
			// Verify file was created
			if _, err := os.Stat(outputFile); os.IsNotExist(err) {
				t.Fatal("Output file was not created")
			}
			
			// Verify file content
			data, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}
			
			// Verify size matches
			if size != len(data) {
				t.Errorf("Expected size %d, got %d", len(data), size)
			}
			
			// Verify it's valid JSON
			var policy testPolicy
			err = json.Unmarshal(data, &policy)
			if err != nil {
				t.Fatalf("Output is not valid JSON: %v", err)
			}
			
			// Verify structure
			if policy.Version != config.SCPVersion {
				t.Errorf("Expected version %s, got %s", config.SCPVersion, policy.Version)
			}
			
			if len(policy.Statement) != len(tt.statements) {
				t.Errorf("Expected %d statements, got %d", len(tt.statements), len(policy.Statement))
			}
			
			// Verify whitespace formatting
			content := string(data)
			hasWhitespace := strings.Contains(content, "\n  ")
			if tt.userInput.Whitespace && !hasWhitespace {
				t.Error("Expected whitespace formatting")
			}
			if !tt.userInput.Whitespace && hasWhitespace {
				t.Error("Expected no whitespace formatting")
			}
		})
	}
}

func TestWriteAllPolicyFiles(t *testing.T) {
	tests := []struct {
		name        string
		userInput   inputs.UserInput
		packedFiles [][]Statement
		outputDir   string
		expected    int
	}{
		{
			name: "Single file output",
			userInput: inputs.UserInput{
				Whitespace: false,
			},
			packedFiles: [][]Statement{
				{
					{Content: map[string]interface{}{"Effect": "Allow"}, Size: 50},
					{Content: map[string]interface{}{"Effect": "Deny"}, Size: 50},
				},
			},
			outputDir: "test",
			expected:  1,
		},
		{
			name: "Multiple file output",
			userInput: inputs.UserInput{
				Whitespace: true,
			},
			packedFiles: [][]Statement{
				{
					{Content: map[string]interface{}{"Effect": "Allow"}, Size: 50},
				},
				{
					{Content: map[string]interface{}{"Effect": "Deny"}, Size: 50},
				},
			},
			outputDir: "test",
			expected:  2,
		},
		{
			name: "Empty packed files",
			userInput: inputs.UserInput{
				Whitespace: false,
			},
			packedFiles: [][]Statement{},
			outputDir:   "test",
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			outputDir := filepath.Join(tempDir, tt.outputDir)
			err := os.MkdirAll(outputDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create output directory: %v", err)
			}
			
			// Create mock input files for testing
			inputFiles := []string{filepath.Join(outputDir, "input.json")}
			results := writeAllPolicyFiles(tt.userInput, tt.packedFiles, outputDir, inputFiles)
			
			if len(results) != tt.expected {
				t.Errorf("Expected %d results, got %d", tt.expected, len(results))
			}
			
			// Verify files were created
			for i, result := range results {
				if _, err := os.Stat(result.Filename); os.IsNotExist(err) {
					t.Errorf("Output file %d was not created: %s", i, result.Filename)
				}
				
				// Verify filename format
				expectedFilename := filepath.Join(outputDir, "corset"+string(rune('1'+i))+".json")
				if result.Filename != expectedFilename {
					t.Errorf("Expected filename %s, got %s", expectedFilename, result.Filename)
				}
				
				// Verify statement count
				if i < len(tt.packedFiles) {
					expectedStatements := len(tt.packedFiles[i])
					if result.Statements != expectedStatements {
						t.Errorf("Expected %d statements in result %d, got %d", expectedStatements, i, result.Statements)
					}
				}
				
				// Verify size is reasonable
				if result.Size <= 0 {
					t.Errorf("Expected positive size for result %d, got %d", i, result.Size)
				}
			}
		})
	}
}

func TestReportResults(t *testing.T) {
	// This is a bit tricky to test since it prints to stdout
	// We'll test that it doesn't panic and basic validation
	tests := []struct {
		name    string
		results []WriteResult
	}{
		{
			name: "Single result",
			results: []WriteResult{
				{
					Filename:   "/tmp/corset1.json",
					Size:       150,
					Statements: 2,
				},
			},
		},
		{
			name: "Multiple results",
			results: []WriteResult{
				{
					Filename:   "/tmp/corset1.json",
					Size:       150,
					Statements: 2,
				},
				{
					Filename:   "/tmp/corset2.json",
					Size:       100,
					Statements: 1,
				},
			},
		},
		{
			name:    "Empty results",
			results: []WriteResult{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify it doesn't panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("reportResults panicked: %v", r)
				}
			}()
			
			reportResults(tt.results)
		})
	}
}

func TestReplaceInputFiles(t *testing.T) {
	tests := []struct {
		name         string
		userInput    inputs.UserInput
		shouldReplace bool
	}{
		{
			name: "Replace single file",
			userInput: inputs.UserInput{
				Replace:     true,
				IsDirectory: false,
			},
			shouldReplace: true,
		},
		{
			name: "Replace directory files",
			userInput: inputs.UserInput{
				Replace:     true,
				IsDirectory: true,
			},
			shouldReplace: true,
		},
		{
			name: "No replace",
			userInput: inputs.UserInput{
				Replace:     false,
				IsDirectory: false,
			},
			shouldReplace: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			testFile := filepath.Join(tempDir, "test.json")
			
			// Create test file
			err := os.WriteFile(testFile, []byte(`{"test": true}`), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
			
			inputFiles := []string{testFile}
			
			// Test the function
			replaceInputFiles(tt.userInput, inputFiles)
			
			// Check if file still exists
			_, err = os.Stat(testFile)
			fileExists := !os.IsNotExist(err)
			
			if tt.shouldReplace && fileExists {
				t.Error("Expected file to be replaced (deleted), but it still exists")
			}
			if !tt.shouldReplace && !fileExists {
				t.Error("Expected file to exist, but it was replaced (deleted)")
			}
		})
	}
}

func TestWriteOutputFiles(t *testing.T) {
	tests := []struct {
		name        string
		userInput   inputs.UserInput
		packedFiles [][]Statement
		inputFiles  []string
	}{
		{
			name: "Complete workflow test",
			userInput: inputs.UserInput{
				Replace:      false,
				Whitespace:  false,
				IsDirectory: false,
			},
			packedFiles: [][]Statement{
				{
					{Content: map[string]interface{}{"Effect": "Allow", "Action": "s3:GetObject"}, Size: 50},
					{Content: map[string]interface{}{"Effect": "Deny", "Action": "s3:DeleteObject"}, Size: 50},
				},
			},
			inputFiles: []string{"/tmp/input.json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			
			// Create a mock input file in the temp directory
			inputFile := filepath.Join(tempDir, "input.json")
			err := os.WriteFile(inputFile, []byte(`{"test": true}`), 0644)
			if err != nil {
				t.Fatalf("Failed to create input file: %v", err)
			}
			tt.inputFiles = []string{inputFile}
			
			// Test the function - should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("writeOutputFiles panicked: %v", r)
				}
			}()
			
			writeOutputFiles(tt.userInput, tt.packedFiles, tt.inputFiles)
			
			// Verify output files were created
			for i := range tt.packedFiles {
				outputFile := filepath.Join(tempDir, "corset"+string(rune('1'+i))+".json")
				if _, err := os.Stat(outputFile); os.IsNotExist(err) {
					t.Errorf("Expected output file %s to be created", outputFile)
				}
			}
		})
	}
}