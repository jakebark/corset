package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindJSONFilesInDirectory(t *testing.T) {
	tests := []struct {
		name          string
		files         map[string]string // filename -> content
		expectedCount int
	}{
		{
			name: "directory with JSON files",
			files: map[string]string{
				"policy1.json": `{"Version": "2012-10-17", "Statement": []}`,
				"policy2.json": `{"Version": "2012-10-17", "Statement": []}`,
				"readme.txt":   "not json",
				"config.yaml":  "also not json",
			},
			expectedCount: 2,
		},
		{
			name: "directory with no JSON files",
			files: map[string]string{
				"readme.txt":  "not json",
				"config.yaml": "also not json",
			},
			expectedCount: 0,
		},
		{
			name:          "empty directory",
			files:         map[string]string{},
			expectedCount: 0,
		},
		{
			name: "directory with nested JSON files",
			files: map[string]string{
				"policy.json":        `{"Version": "2012-10-17"}`,
				"subdir/nested.json": `{"Version": "2012-10-17"}`,
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create test files
			for filename, content := range tt.files {
				filePath := filepath.Join(tempDir, filename)

				// Create directory if needed
				dir := filepath.Dir(filePath)
				if dir != tempDir {
					err := os.MkdirAll(dir, 0755)
					if err != nil {
						t.Fatalf("Failed to create directory %s: %v", dir, err)
					}
				}

				err := os.WriteFile(filePath, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file %s: %v", filePath, err)
				}
			}

			// Test the function
			result := FindJSONFilesInDirectory(tempDir)

			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d JSON files, got %d", tt.expectedCount, len(result))
			}

			// Verify all returned files are JSON files
			for _, file := range result {
				if !filepath.IsAbs(file) {
					t.Errorf("Expected absolute path, got %s", file)
				}
				if filepath.Ext(file) != ".json" {
					t.Errorf("Expected .json extension, got %s", file)
				}
			}
		})
	}
}

func TestIsDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.json")
	err := os.WriteFile(testFile, []byte(`{}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Directory path",
			path:     tempDir,
			expected: true,
		},
		{
			name:     "File path",
			path:     testFile,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: We'll need to move isDirectory from inputs package or create a local version
			// For now, testing the concept with os.Stat directly
			info, err := os.Stat(tt.path)
			if err != nil {
				t.Fatalf("Failed to stat path %s: %v", tt.path, err)
			}

			result := info.IsDir()
			if result != tt.expected {
				t.Errorf("Expected IsDirectory(%s) = %v, got %v", tt.path, tt.expected, result)
			}
		})
	}
}

