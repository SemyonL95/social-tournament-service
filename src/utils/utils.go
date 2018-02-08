package utils

import (
	"regexp"
	"fmt"
)

type NotFoundError struct {
	Text string
}

type ForbiddenError struct {
	Text string
}

func (err *ForbiddenError) Error() string {
	return fmt.Sprintf("%s %s", err.Text, "Forbidden")
}

func (err *NotFoundError) Error() string {
	return fmt.Sprintf("%s %s", err.Text, "Not Found")
}

func ValidateString (value string) (bool) {
	if m, _ := regexp.MatchString("^[A-Za-z0-9]{1,20}$", value); !m {
		return false
	}

	return true
}

func ValidateFloatNotNagtive (value float64) (bool) {
	if value < 0 {
		return false
	}

	return true
}