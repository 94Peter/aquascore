package crawler

type Option func(*ctsaCrawler) error

func WithBaseURL(baseURL string) Option {
	return func(c *ctsaCrawler) error {
		c.baseUrl = baseURL
		return nil
	}
}

func WithPersistence(persistence Persistence) Option {
	return func(c *ctsaCrawler) error {
		c.persistence = persistence
		return nil
	}
}

func withTest(test bool) Option {
	return func(c *ctsaCrawler) error {
		c.isTest = test
		return nil
	}
}
