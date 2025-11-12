package utils

import "strings"

// ScopeContains checks if a given scope string contains a specific scope.
func ScopeContains(scope, s string) bool {
	scopes := strings.Split(scope, " ")
	for _, sc := range scopes {
		if sc == s {
			return true
		}
	}
	return false
}
