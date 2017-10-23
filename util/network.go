package util

import "strconv"

func JoinURL(parts ...interface{}) string {
	url := ""
	for _, p := range parts {
		if _, ok := p.(string); ok {
			url = url + "/" + p.(string)
		}
		if _, ok := p.(int); ok {
			url = url + "/" + strconv.Itoa(p.(int))
		}
	}
	return url
}
