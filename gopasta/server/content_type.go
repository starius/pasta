package server

import "strings"

func contentTypeIsInline(ctype string) bool {
	if strings.HasPrefix(ctype, "text/") {
		return true
	}
	if strings.HasPrefix(ctype, "image/png") {
		return true
	}
	if strings.HasPrefix(ctype, "image/jpeg") {
		return true
	}
	if strings.HasPrefix(ctype, "image/gif") {
		return true
	}
	if strings.HasPrefix(ctype, "audio/mpeg") {
		return true
	}
	if strings.HasPrefix(ctype, "audio/ogg") {
		return true
	}
	if strings.HasPrefix(ctype, "application/json") {
		return true
	}
	if strings.HasPrefix(ctype, "application/javascript") {
		return true
	}
	return false
}
