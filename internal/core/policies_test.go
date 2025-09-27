package core

import (
	"testing"

	"github.com/jakebark/corset/internal/config"
)

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

