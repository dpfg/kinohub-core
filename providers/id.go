package providers

import (
	"strings"
	"unicode"
)

const (
	// IDTypeKinoHub - kino.pub
	IDTypeKinoHub = "KH"
	// IDTypeTMDB - tmdb.com
	IDTypeTMDB = "TM"
	// IDTypeTrakt - trakt.tv
	IDTypeTrakt = "TK"
	// IDTypeSeasonvar - seasonvar.ru
	IDTypeSeasonvar = "SV"
)

// MatchUIDType check is the uid matches to provided id type
func MatchUIDType(uid, idType string) bool {
	ni := strings.IndexFunc(strings.ToUpper(uid), func(r rune) bool {
		return unicode.IsNumber(r)
	})

	if ni >= len(uid) || ni < 0 {
		return false
	}

	return idType == uid[0:ni]
}
