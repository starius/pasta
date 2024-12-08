package blockedbots

import "testing"

func TestIsBlockedCrawler(t *testing.T) {
	cases := []struct {
		userAgent   string
		wantBlocked bool
	}{
		{
			userAgent:   "LinkedInBot/1.0 (compatible; Mozilla/5.0; Apache-HttpClient +http://www.linkedin.com)",
			wantBlocked: true,
		},
		{
			userAgent:   "wget",
			wantBlocked: false,
		},
		{
			userAgent:   "curl/7.29.0",
			wantBlocked: false,
		},
		{
			userAgent:   "PageThing http://pagething.com curl www",
			wantBlocked: true,
		},
		{
			userAgent:   "PHP-Curl-Class/4.13.0 (+https://github.com/php-curl-class/php-curl-class) PHP/7.4.11 curl/7.69.1",
			wantBlocked: true,
		},
		{
			userAgent:   "serpstatbot/1.0 (advanced backlink tracking bot; curl/7.58.0; http://serpstatbot.com/; abuse@serpstatbot.com)",
			wantBlocked: true,
		},
		{
			userAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1.1 Safari/605.1.15",
			wantBlocked: false,
		},
		{
			userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:132.0) Gecko/20100101 Firefox/132.0",
			wantBlocked: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.userAgent, func(t *testing.T) {
			got := IsBlockedCrawler(tc.userAgent)
			if got != tc.wantBlocked {
				t.Errorf("mismatch for %q: want %v, got %v", tc.userAgent, tc.wantBlocked, got)
			}
		})
	}
}
