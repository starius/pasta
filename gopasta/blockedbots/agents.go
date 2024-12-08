package blockedbots

import (
	agents "github.com/monperrus/crawler-user-agents"
)

var curlBotIndex = func() int {
	for i, crawler := range agents.Crawlers {
		if crawler.URL == "https://curl.haxx.se/" {
			return i
		}
	}

	panic("can't find CURL crawler")
}()

var wgetBotIndex = func() int {
	for i, crawler := range agents.Crawlers {
		if crawler.Pattern == "[wW]get" {
			return i
		}
	}

	panic("can't find WGET crawler")
}()

// IsBlockedCrawler returns if the user agent belongs to a blocked bot.
// All bots are blocked, except curl and wget, because they are used manually.
func IsBlockedCrawler(userAgent string) bool {
	// Find all matches.
	matches := agents.MatchingCrawlers(userAgent)

	// Filter out CURL and WGET, because they can be used manually.
	for _, m := range matches {
		if m == curlBotIndex || m == wgetBotIndex {
			continue
		}

		return true
	}

	return false
}
