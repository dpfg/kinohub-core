package util

func PadLeft(str, pad string, lenght int) string {
	for {
		if len(str) >= lenght {
			return str
		}
		str = pad + str
	}
}
