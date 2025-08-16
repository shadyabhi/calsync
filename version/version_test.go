package version

import (
	"testing"
)

func TestIsDifferent(t *testing.T) {
	tests := []struct {
		name     string
		version1 string
		version2 string
		expected bool
	}{
		{
			name:     "different versions",
			version1: "2.0.0",
			version2: "1.0.0",
			expected: true,
		},
		{
			name:     "same version",
			version1: "1.1.1",
			version2: "1.1.1",
			expected: false,
		},
		{
			name:     "version with commit hash - different",
			version1: "1.2.0-abc123-2024-01-01",
			version2: "1.1.0-def456-2024-01-01",
			expected: true,
		},
		{
			name:     "version with commit hash - same base version",
			version1: "1.2.0-abc123-2024-01-01",
			version2: "1.2.0-def456-2024-01-02",
			expected: false,
		},
		{
			name:     "different patch versions",
			version1: "0.1.16",
			version2: "0.1.14",
			expected: true,
		},
		{
			name:     "identical versions",
			version1: "2.1.3",
			version2: "2.1.3",
			expected: false,
		},
	}

	checker := New("1.0.0")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.isDifferent(tt.version1, tt.version2)
			if result != tt.expected {
				t.Errorf("isDifferent(%s, %s) = %v, want %v", tt.version1, tt.version2, result, tt.expected)
			}
		})
	}
}

func TestCleanVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "version with commit and date",
			input:    "1.2.3-abc123-2024-01-01",
			expected: "1.2.3",
		},
		{
			name:     "clean version",
			input:    "1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "version with only commit",
			input:    "1.2.3-abc123",
			expected: "1.2.3",
		},
	}

	checker := New("1.0.0")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.cleanVersion(tt.input)
			if result != tt.expected {
				t.Errorf("cleanVersion(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCheckForUpdateDevVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "exact dev version",
			version: "dev",
		},
		{
			name:    "version containing dev",
			version: "1.2.3-dev",
		},
		{
			name:    "version with dev suffix",
			version: "2.0.0-dev-abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := New(tt.version)

			result := make(chan *UpdateInfo, 1)
			checker.CheckForUpdate(result)

			updateInfo := <-result
			if updateInfo != nil {
				t.Errorf("CheckForUpdate() with version %s should return nil, got %+v", tt.version, updateInfo)
			}
		})
	}
}
