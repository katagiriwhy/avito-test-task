package validation

import "strings"

func IsEmptyString(s string) bool {
	return strings.TrimSpace(s) == ""
}
