package blockedbots

import (
	agents "github.com/monperrus/crawler-user-agents"
)

// IsBlockedCrawler returns if the user agent belongs to a blocked bot.
func IsBlockedCrawler(userAgent string) bool {
	return agents.IsCrawler(userAgent)
}
