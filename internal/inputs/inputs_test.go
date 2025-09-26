package inputs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jakebark/corset/internal/config"
)

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
			result := isDirectory(tt.path)
			if result != tt.expected {
				t.Errorf("Expected isDirectory(%s) = %v, got %v", tt.path, tt.expected, result)
			}
		})
	}
}

func TestUserInputStructure(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		delete      bool
		whitespace  bool
		isDirectory bool
		maxFiles    int
	}{
		{
			name:        "File input configuration",
			target:      "/path/to/file.json",
			delete:      false,
			whitespace:  true,
			isDirectory: false,
			maxFiles:    config.DefaultMaxFiles,
		},
		{
			name:        "Directory input configuration",
			target:      "/path/to/directory",
			delete:      true,
			whitespace:  false,
			isDirectory: true,
			maxFiles:    config.DefaultMaxFiles,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userInput := UserInput{
				Target:      tt.target,
				Delete:      tt.delete,
				Whitespace:  tt.whitespace,
				IsDirectory: tt.isDirectory,
				MaxFiles:    tt.maxFiles,
			}
			
			if userInput.Target != tt.target {
				t.Errorf("Expected Target %s, got %s", tt.target, userInput.Target)
			}
			if userInput.Delete != tt.delete {
				t.Errorf("Expected Delete %v, got %v", tt.delete, userInput.Delete)
			}
			if userInput.Whitespace != tt.whitespace {
				t.Errorf("Expected Whitespace %v, got %v", tt.whitespace, userInput.Whitespace)
			}
			if userInput.IsDirectory != tt.isDirectory {
				t.Errorf("Expected IsDirectory %v, got %v", tt.isDirectory, userInput.IsDirectory)
			}
			if userInput.MaxFiles != tt.maxFiles {
				t.Errorf("Expected MaxFiles %d, got %d", tt.maxFiles, userInput.MaxFiles)
			}
		})
	}
}

// Note: Testing ParseFlags() would require mocking command line arguments
// which is more complex and might be better suited for integration tests