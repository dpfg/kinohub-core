package util

import (
	"strconv"
)

func JoinURL(parts ...interface{}) string {
	url := ""
	for _, p := range parts {
		if _, ok := p.(string); ok {
			url = append(url, p.(string))
		}
		if _, ok := p.(int); ok {
			url = append(url, strconv.Itoa(p.(int)))
		}
	}
	return url
}

func append(p1, p2 string) string {
	if p1 == "" {
		return p1 + p2
	}
	return p1 + "/" + p2
}
