package utils

import (
	"fmt"
	"math"
	"strconv"

	"github.com/Kameleoon/client-go/v3/errs"
	"github.com/Kameleoon/client-go/v3/logging"
)

// Version represents a semantic version with major, minor, and patch components.
type Version struct {
	Major int
	Minor int
	Patch int
}

// NewVersionFromString parses a version string (e.g., "1.2.3") and returns a Version if no parsing fails.
// A pre-release (-) or build (+) suffix is ignored (e.g., "1.2.3-beta.1+build" is parsed as "1.2.3").
func NewVersionFromString(versionString string) (*Version, error) {
	versions := [3]int{}
	index := 0
	hasDigits := false
	valid := true
	// Single pass without splitting, stops at pre-release (-) or build (+) suffix
	for i := 0; i < len(versionString); i++ {
		c := versionString[i]
		if c >= '0' && c <= '9' && versions[index] <= (math.MaxInt-int(c-'0'))/10 {
			versions[index] = versions[index]*10 + int(c-'0')
			hasDigits = true
		} else if c == '.' && hasDigits && index < 2 {
			index++
			hasDigits = false
		} else {
			valid = c == '.' || c == '-' || c == '+'
			break
		}
	}
	// Each parsed component must have at least one digit
	if !valid || !hasDigits {
		logging.Error("Invalid version component, index: %d, version: '%s'", index, versionString)
		return nil, errs.NewInternalError("Parsing error")
	}
	return &Version{Major: versions[0], Minor: versions[1], Patch: versions[2]}, nil
}

// CompareTo compares this version with another.
// Returns: -1 if this < other, 0 if equal, 1 if this > other.
func (v *Version) CompareTo(other *Version) int {
	if cmp := compareInt(v.Major, other.Major); cmp != 0 {
		return cmp
	}
	if cmp := compareInt(v.Minor, other.Minor); cmp != 0 {
		return cmp
	}
	return compareInt(v.Patch, other.Patch)
}

// ToFloat returns version as a float32 (major, minor)
func (v *Version) ToFloat() (float32, error) {
	versionString := fmt.Sprintf("%d.%d", v.Major, v.Minor)
	if versionFloat, err := strconv.ParseFloat(versionString, 32); err == nil {
		return float32(versionFloat), nil
	}
	logging.Error(fmt.Sprintf("Failed to parse Version: %v", v))
	return 0.0, errs.NewInternalError(fmt.Sprintf("ToFloat parsing failed: %v", v))
}

func compareInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
