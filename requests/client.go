package requests

import "github.com/gocolly/colly"

type Client interface {
	CallEndpoint(c *colly.Collector, args ...interface{}) *colly.Collector
}
