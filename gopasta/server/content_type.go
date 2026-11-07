package server

import "mime"

// Only passive media and plain text are safe to display as uploaded content.
func contentTypeIsInline(ctype string) bool {
	mediaType, _, err := mime.ParseMediaType(ctype)
	if err != nil {
		return false
	}
	switch mediaType {
	case "text/plain", "image/png", "image/jpeg", "image/gif", "image/webp",
		"audio/mpeg", "audio/ogg", "application/json":
		return true
	default:
		return false
	}
}
