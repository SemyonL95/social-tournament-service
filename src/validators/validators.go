package validators

import (
	"regexp"
)

func ValidateUsername(value string) (bool) {
	if m, _ := regexp.MatchString("^[A-Za-z0-9]{1,20}$", value); !m {
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
