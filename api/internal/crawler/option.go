package crawler

import "io"

type Option func(*ctsaCrawler)

func WithBaseURL(baseURL string) Option {
	return func(c *ctsaCrawler) {
		c.baseUrl = baseURL
	}
}

func WithPersistence(persistence Persistence) Option {
	return func(c *ctsaCrawler) {
		c.persistence = persistence
	}
}

func withGetResponse(mock func(url string) (io.Reader, error)) Option {
	return func(c *ctsaCrawler) {
		c.mockGetResponse = mock
	}
}
