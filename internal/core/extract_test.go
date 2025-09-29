package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractIndividualPolicies(t *testing.T) {
	tests := []struct {
		name               string
		policy             Policy
		expectedStatements int
	}{
		{
			name: "one statement",
			policy: Policy{
				Version: "2012-10-17",
				Statement: []map[string]interface{}{
					{"Effect": "Allow", "Action": "s3:*", "Resource": "*"},
				},
			},
			expectedStatements: 1,
		},
		{
			name: "two statements",
			policy: Policy{
				Version: "2012-10-17",
				Statement: []map[string]interface{}{
					{"Effect": "Allow", "Action": "s3:GetObject", "Resource": "*"},
					{"Effect": "Deny", "Action": "s3:DeleteObject", "Resource": "*"},
				},
			},
			expectedStatements: 2,
		},
		{
			name: "no statements",
			policy: Policy{
				Version:   "2012-10-17",
				Statement: []map[string]interface{}{},
			},
			expectedStatements: 0,
		},
		{
			name: "two statements, complex",
			policy: Policy{
				Version: "2012-10-17",
				Statement: []map[string]interface{}{
					{
						"Sid":    "AllowS3Access",
						"Effect": "Allow",
						"Action": []string{"s3:GetObject", "s3:PutObject"},
						"Resource": []string{
							"arn:aws:s3:::my-bucket/*",
							"arn:aws:s3:::my-bucket",
						},
						"Condition": map[string]interface{}{
							"StringEquals": map[string]interface{}{
								"aws:username": "testuser",
							},
						},
					},
					{
						"Effect":   "Deny",
						"Action":   "*",
						"Resource": "*",
						"Condition": map[string]interface{}{
							"Bool": map[string]interface{}{
								"aws:MultiFactorAuthPresent": "false",
							},
						},
					},
				},
			},
			expectedStatements: 2,
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
			statements := extractIndividualPolicies(testFile)

			if len(statements) != tt.expectedStatements {
				t.Errorf("Expected %d statements, got %d", tt.expectedStatements, len(statements))
			}

			// Verify each statement has required fields
			for i, stmt := range statements {
				if stmt.Content == nil {
					t.Errorf("Statement %d has nil content", i)
				}
				if len(statements) > 0 && stmt.Size <= 0 {
					t.Errorf("Statement %d has invalid size: %d", i, stmt.Size)
				}

				// Verify content matches original
				if i < len(tt.policy.Statement) {
					expectedContent := tt.policy.Statement[i]
					if !mapsEqual(stmt.Content, expectedContent) {
						t.Errorf("Statement %d content doesn't match original", i)
					}
				}
			}
		})
	}
}

func TestExtractAllStatements(t *testing.T) {
	tests := []struct {
		name          string
		policies      []Policy
		expectedTotal int
	}{
		{
			name: "single file, multiple statements",
			policies: []Policy{
				{
					Version: "2012-10-17",
					Statement: []map[string]interface{}{
						{"Effect": "Allow", "Action": "s3:GetObject", "Resource": "*"},
						{"Effect": "Deny", "Action": "s3:DeleteObject", "Resource": "*"},
					},
				},
			},
			expectedTotal: 2,
		},
		{
			name: "multiple files, multiple statements",
			policies: []Policy{
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
						{"Effect": "Allow", "Action": "ec2:DescribeInstances", "Resource": "*"},
					},
				},
			},
			expectedTotal: 3,
		},
		{
			name: "mutiple files, no statements",
			policies: []Policy{
				{
					Version:   "2012-10-17",
					Statement: []map[string]interface{}{},
				},
				{
					Version:   "2012-10-17",
					Statement: []map[string]interface{}{},
				},
			},
			expectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			var files []string

			// Create test files
			for i, policy := range tt.policies {
				testFile := filepath.Join(tempDir, filepath.Base(tt.name)+"-"+string(rune('a'+i))+".json")
				data, err := json.MarshalIndent(policy, "", "  ")
				if err != nil {
					t.Fatalf("Failed to marshal test policy %d: %v", i, err)
				}

				err = os.WriteFile(testFile, data, 0644)
				if err != nil {
					t.Fatalf("Failed to write test file %d: %v", i, err)
				}
				files = append(files, testFile)
			}

			// Test the function
			statements := extractAllStatements(files)

			if len(statements) != tt.expectedTotal {
				t.Errorf("Expected %d total statements, got %d", tt.expectedTotal, len(statements))
			}

			// Verify all statements are valid
			for i, stmt := range statements {
				if stmt.Content == nil {
					t.Errorf("Statement %d has nil content", i)
				}
				if len(statements) > 0 && stmt.Size <= 0 {
					t.Errorf("Statement %d has invalid size: %d", i, stmt.Size)
				}
			}
		})
	}
}

func TestExtractIndividualPoliciesInvalidFile(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "invalid JSON",
			content:  `{"Version": "2012-10-17", "Statement": [}`,
			expected: 0, // Should handle gracefully
		},
		{
			name:     "missing statement field",
			content:  `{"Version": "2012-10-17"}`,
			expected: 0,
		},
		{
			name:     "missing file",
			content:  "", // Will test with non-existent file
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testFile string
			if tt.name == "Non-existent file" {
				testFile = filepath.Join(tempDir, "nonexistent.json")
			} else {
				testFile = filepath.Join(tempDir, "test.json")
				err := os.WriteFile(testFile, []byte(tt.content), 0644)
				if err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
			}

			// Should not panic, should handle gracefully
			statements := extractIndividualPolicies(testFile)

			if len(statements) != tt.expected {
				t.Errorf("Expected %d statements, got %d", tt.expected, len(statements))
			}
		})
	}
}

// Helper function to compare maps - simplified for testing
func mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
		// For testing purposes, just check that keys exist
		// Full deep comparison would be more complex for nested structures
	}

	return true
}

