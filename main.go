package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type Crawler struct {
	startUrl            string
	depth               int
	visitedDomains      map[string]struct{}
	mutexVisitedDomains sync.Mutex
	mu                  sync.Mutex
}

func NewCrawler(url string, depth int) Crawler {
	return Crawler{startUrl: url, depth: depth, visitedDomains: make(map[string]struct{})}
}

func (c *Crawler) crawl() {
	urlsToFetch := []string{c.startUrl}
	startUrlDomain := strings.Split(c.startUrl, "/")[2]
	c.visitedDomains[startUrlDomain] = struct{}{}
	var nextUrlsToFetch []string
	var wg sync.WaitGroup
	for i := 1; i <= c.depth; i++ {
		// fmt.Println("******* ", urlsToFetch)
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

func (c *Crawler) fetchUniqueUrlsOnPage(url string, wg *sync.WaitGroup) []string {
	defer wg.Done()

	urls := findUrlsOnPage(url)
	var uniqueUrls []string
	for _, u := range urls {
		domain := strings.Split(u, "/")[2]
		c.mutexVisitedDomains.Lock()
		_, exist := c.visitedDomains[domain]
		c.mutexVisitedDomains.Unlock()
		if !exist {
			c.mutexVisitedDomains.Lock()
			c.visitedDomains[domain] = struct{}{}
			c.mutexVisitedDomains.Unlock()
			uniqueUrls = append(uniqueUrls, u)
		}
	}
	return uniqueUrls
}

func findUrlsOnPage(url string) []string {
	htmlDoc := getHtmlParsedResponse(url)
	return traverseHtmlDoc(htmlDoc)
}

func traverseHtmlDoc(node *html.Node) []string {
	if node == nil {
		return []string{}
	}
	var urls []string

	if node.Type == html.ElementNode && node.Data == "a" {
		for _, attr := range node.Attr {
			if attr.Key == "href" && (strings.HasPrefix(attr.Val, "https://") || strings.HasPrefix(attr.Val, "http://")) {
				urls = append(urls, attr.Val)
			}
		}

	}

	for childNode := node.FirstChild; childNode != nil; childNode = childNode.NextSibling {
		childNodeUrls := traverseHtmlDoc(childNode)
		urls = append(urls, childNodeUrls...)
	}

	return urls
}

func getHtmlParsedResponse(url string) *html.Node {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating new http request. Error: ", err)
		return nil
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error fetching http request's response. Error: ", err)
		return nil
	}

	defer resp.Body.Close()

	htmlDoc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Println("Error parsing html response body. Error: ", err)
		return nil
	}
	return htmlDoc
}

func main() {
	// url := "https://example.com"
	url := "https://go.dev/doc/articles/wiki/"
	// url := "https://jvns.ca/blog/2017/04/01/slow-down-your-internet-with-tc/"
	depth := 3
	c := NewCrawler(url, depth)
	c.crawl()
	fmt.Println("Crawled domains: ")
	for domain := range c.visitedDomains {
		fmt.Println(domain)
	}
}
