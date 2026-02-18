package versionutils

import (
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/version"
)

type Version struct {
	Major uint64
	Minor uint64
}

func (v Version) String() string {
	return fmt.Sprintf("v%d.%d.X", v.Major, v.Minor)
}

var DRASupportMinStable = Version{1, 34}

func HasStableDRASupport(version *version.Info) (bool, error) {
	return IsAtleast(version, DRASupportMinStable)
}

func IsAtleast(version *version.Info, min Version) (bool, error) {
	major, err := strconv.ParseUint(version.Major, 10, 64)
	if err != nil {
		return false, err
	}
	minor, err := strconv.ParseUint(version.Minor, 10, 64)
	if err != nil {
		return false, err
	}

	if major < min.Major {
		return false, nil
	} else if major == min.Major && minor < min.Minor {
		return false, nil
	}

	return true, nil
}
