package validators

import (
	"fmt"
	"net/http"
	"regexp"
)

var playerIDRegexp = regexp.MustCompile("^[A-Za-z0-9]{1,20}$")

func ValidateUsername(value string, naming string, w http.ResponseWriter) bool {
	if m := playerIDRegexp.MatchString(value); !m {
		msg := fmt.Sprintf("%s have to be a string A-Za-z0-9 min: 1, max: 20 characters", naming)
		http.Error(w, msg, http.StatusUnprocessableEntity)
		return false
	}

	return true
}

func ValidateFloatNotNegative(value float64, naming string, w http.ResponseWriter) bool {
	if value < 0 {
		msg := fmt.Sprintf("%s is required and points have to be numeric and not negative", naming)
		http.Error(w, msg, http.StatusUnprocessableEntity)
		return false
	}

	return true
}

func ValidateIntNotNegative(value int, naming string, w http.ResponseWriter) bool {
	if value < 0 {
		msg := fmt.Sprintf("%s is required and points have to be numeric and not negative", naming)
		http.Error(w, msg, http.StatusUnprocessableEntity)
		return false
	}

	return true
}

func ValidateMethod(meth string, r *http.Request, w http.ResponseWriter) bool {
	if r.Method != meth {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}
