package validators

import (
	"regexp"
)

var playerIDRegexp = regexp.MustCompile("^[A-Za-z0-9]{1,20}$")

func ValidateUsername(value string) (bool) {
	if m := playerIDRegexp.MatchString(value); !m {
		return false
	}

	return true
}

func ValidateFloatNotNegative(value float64) (bool) {
	if value < 0 {
		return false
	}

	return true
}
