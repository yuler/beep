package updater

import "testing"

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"0.2.1", "0.2.1", 0},
		{"v0.2.1", "0.2.1", 0},
		{"0.2.1", "v0.2.1", 0},
		{"v0.2.2", "v0.2.1", 1},
		{"v0.2.0", "v0.2.1", -1},
		{"v1.0.0", "v0.9.9", 1},
		{"v0.10.0", "v0.9.9", 1},
		{"v0.2.1-rc1", "v0.2.1", -1},
		{"v0.2.1", "v0.2.1-rc1", 1},
		{"v0.2.1-rc1", "v0.2.1-rc2", -1},
		{"dev", "v0.2.1", -1},
		{"v0.2.1", "dev", 1},
		{"dev", "dev", 0},
	}

	for _, tt := range tests {
		actual := CompareVersions(tt.v1, tt.v2)
		if actual != tt.expected {
			t.Errorf("CompareVersions(%q, %q) = %d; want %d", tt.v1, tt.v2, actual, tt.expected)
		}
	}
}
