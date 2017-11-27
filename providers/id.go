package providers

import (
	"strings"
	"unicode"
)

const (
	ID_TYPE_KINOHUB = "KH"
	ID_TYPE_TMDB    = "TM"
	ID_TYPE_TRAKT   = "TK"
)

func MatchUIDType(uid, idType string) bool {
	ni := strings.IndexFunc(strings.ToUpper(uid), func(r rune) bool {
		return unicode.IsNumber(r)
	})

	if ni >= len(uid) || ni < 0 {
		return false
	}

	return idType == uid[0:ni]
}
