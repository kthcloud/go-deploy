package versionutils

import (
	"fmt"
	"regexp"
	"testing"

	"k8s.io/apimachinery/pkg/version"
)

func TestIsAtleast(t *testing.T) {
	tests := []struct {
		name     string
		version  *version.Info
		minimum  Version
		expected bool
	}{
		{
			name:     "Version matches exactly",
			version:  &version.Info{Major: "1", Minor: "34"},
			minimum:  DRASupportMinStable,
			expected: true,
		},
		{
			name:     "Version is greater than minimum",
			version:  &version.Info{Major: "1", Minor: "35"},
			minimum:  DRASupportMinStable,
			expected: true,
		},
		{
			name:     "Version is less than minimum (minor)",
			version:  &version.Info{Major: "1", Minor: "33"},
			minimum:  DRASupportMinStable,
			expected: false,
		},
		{
			name:     "Version is less than minimum (major)",
			version:  &version.Info{Major: "0", Minor: "50"},
			minimum:  DRASupportMinStable,
			expected: false,
		},
		{
			name:     "Version is greater than minimum (major)",
			version:  &version.Info{Major: "2", Minor: "0"},
			minimum:  DRASupportMinStable,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := IsAtleast(tt.version, tt.minimum)
			if got != tt.expected {
				t.Errorf("IsAtleast() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHasStableDRASupport(t *testing.T) {
	testClusterVersion, err := ParseVersion("v1.34.1+rke2r1")
	if err != nil {
		t.Error("failed to parse testClusterVersion:", err)
		return
	}
	tests := []struct {
		name     string
		version  *version.Info
		expected bool
	}{
		{
			name:     "Version matches minimum stable DRA support",
			version:  &version.Info{Major: "1", Minor: "34"},
			expected: true,
		},
		{
			name:     "Version above minimum stable DRA support",
			version:  &version.Info{Major: "1", Minor: "35"},
			expected: true,
		},
		{
			name:     "Version below minimum stable DRA support",
			version:  &version.Info{Major: "1", Minor: "33"},
			expected: false,
		},
		{
			name:     "Version well below minimum stable DRA support",
			version:  &version.Info{Major: "0", Minor: "50"},
			expected: false,
		},
		{
			name:     "Version above major threshold",
			version:  &version.Info{Major: "2", Minor: "0"},
			expected: true,
		},
		{
			name:     "Real k8s version from test cluster",
			version:  testClusterVersion,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := HasStableDRASupport(tt.version)
			if got != tt.expected {
				t.Errorf("HasStableDRASupport() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func ParseVersion(versionStr string) (*version.Info, error) {
	re := regexp.MustCompile(`^v(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)(\+.*)?$`)
	matches := re.FindStringSubmatch(versionStr)
	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid version string: %s", versionStr)
	}

	major := matches[1]
	minor := matches[2]

	return &version.Info{
		Major:      major,
		Minor:      minor,
		GitVersion: versionStr,
	}, nil
}
