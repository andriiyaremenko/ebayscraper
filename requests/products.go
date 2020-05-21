package requests

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/andriiyaremenko/ebayscraper/data"
	"github.com/andriiyaremenko/ebayscraper/types"
	"github.com/andriiyaremenko/tinylog"
	"github.com/gocolly/colly"
)

var (
	pageUrlRegex = regexp.MustCompile(`pgn=\d+$`)
)

func NewProductsClient(loggerFactory tinylog.TinyLoggerFactory, fs data.FileStorage) Client {
	return &productsClient{
		fs:           fs,
		logger:       loggerFactory.GetLogger("ProductsClient"),
		visitedPages: make(map[string]struct{}),
	}
}

type productsClient struct {
	mu           sync.Mutex
	visitedPages map[string]struct{}
	fs           data.FileStorage
	logger       tinylog.TinyLogger
}

// (pc *productsClient) CallEndpoint(c *col.Collector, category string) *colly.Collector
func (pc *productsClient) CallEndpoint(c *colly.Collector, args ...interface{}) *colly.Collector {
	if len(args) == 0 {
		pc.logger.Fatal("(pc *productsClient) CallEndpoint(category string): to few arguments")
	}
	if len(args) > 1 {
		pc.logger.Fatal("(pc *productsClient) CallEndpoint(category string): to much arguments")
	}
	category := args[0].(string)
	start := time.Now()
	c.Async = true

	c.OnResponse(func(r *colly.Response) {
		pc.logger.Debugf("%s: %s%s -> %d", r.Request.Method, r.Request.URL.Hostname(), r.Request.URL.RequestURI(), r.StatusCode)
	})

	c.OnError(func(r *colly.Response, err error) {
		pc.logger.Fatalf("Status Code: %d; %v", r.StatusCode, err)
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		page := pageUrlRegex.FindString(href)
		pc.mu.Lock()
		if _, ok := pc.visitedPages[page]; !ok && strings.Contains(href, category) && pageUrlRegex.MatchString(href) {
			pc.visitedPages[page] = struct{}{}
			pc.mu.Unlock()
			e.Request.Visit(href)
			return
		}
		pc.mu.Unlock()
	})

	c.OnHTML("li.s-item", func(e *colly.HTMLElement) {
		item := new(types.Product)
		item.Title = e.ChildText("h3.s-item__title")
		item.ImageLink = e.ChildAttr("img.s-item__image-img", "src")
		item.Attributes = make(map[string]string)
		e.ForEach("span.s-item__dynamic", func(i int, ch *colly.HTMLElement) {
			if ch.Text == "" {
				return
			}
			keyValue := strings.SplitN(ch.Text, ": ", 2)
			item.Attributes[keyValue[0]] = keyValue[1]
		})
		err := pc.fs.Save(*item)
		if err != nil {
			pc.logger.Fatal(err)
		}
	})

	err := c.Visit(fmt.Sprintf("https://www.ebay.com/%s", category))
	if err != nil {
		pc.logger.Fatal(err)
	}
	c.Wait()
	err = pc.fs.Flush()
	elapsed := time.Since(start)
	pc.logger.Infof("Finished products scrapping in %s", elapsed)
	if err != nil {
		pc.logger.Fatal(err)
	}
	result := c.Clone()
	result.Async = false
	return result
}
