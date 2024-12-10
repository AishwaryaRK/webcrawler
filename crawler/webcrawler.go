package crawler

import (
	"log"
	"net/url"
	"sync"
	"web_crawler/parser"
)

type Crawler struct {
	startUrl       string
	depth          int
	visitedDomains sync.Map
	mu             sync.Mutex
	httpParser     *parser.HttpClient
}

func NewCrawler(url string, depth int, timeout int) *Crawler {
	return &Crawler{startUrl: url, depth: depth, httpParser: parser.NewHttpClient(timeout)}
}

func (c *Crawler) Crawl() {
	urlsToFetch := []string{c.startUrl}
	parsedUrl, err := url.Parse(c.startUrl)
	if err != nil {
		log.Println("Invalid start url: ", c.startUrl)
		return
	}

	startUrlDomain := parsedUrl.Hostname()
	c.visitedDomains.Store(startUrlDomain, struct{}{})
	var nextUrlsToFetch []string
	var wg sync.WaitGroup
	for i := 1; i <= c.depth; i++ {
		wg.Add(len(urlsToFetch))
		for _, url := range urlsToFetch {
			go func() {
				uniqueUrls := c.fetchUniqueUrlsOnPage(url, &wg)
				c.mu.Lock()
				nextUrlsToFetch = append(nextUrlsToFetch, uniqueUrls...)
				c.mu.Unlock()
			}()
		}
		wg.Wait()
		urlsToFetch = nextUrlsToFetch
	}
}

func (c *Crawler) fetchUniqueUrlsOnPage(urlStr string, wg *sync.WaitGroup) []string {
	defer wg.Done()

	urls := c.httpParser.FindUrlsOnPage(urlStr)
	var uniqueUrls []string
	for _, u := range urls {
		parsedUrl, err := url.Parse(u)
		if err != nil {
			log.Println("Invalid start url: ", c.startUrl)
			continue
		}
		domain := parsedUrl.Hostname()
		if _, exist := c.visitedDomains.Load(domain); !exist {
			c.visitedDomains.Store(domain, struct{}{})
			uniqueUrls = append(uniqueUrls, u)
		}
	}
	return uniqueUrls
}

func (c *Crawler) GetVisitedDomains() []string {
	var domains []string
	c.visitedDomains.Range(func(key, value interface{}) bool {
		if domain, ok := key.(string); ok {
			domains = append(domains, domain)
		}
		return true
	})
	return domains
}
