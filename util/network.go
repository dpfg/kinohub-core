package util

import "strings"

func JoinURL(parts ...string) string {
	return strings.Join(parts, "/")
}
