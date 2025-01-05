package main

import (
	"fmt"
	"web_crawler/crawler"
)

func main() {
	// url := "https://example.com"
	//url := "https://go.dev/doc/articles/wiki/"
	url := "https://jvns.ca/blog/2017/04/01/slow-down-your-internet-with-tc/"
	depth := 3
	timeout := 5
	c := crawler.NewCrawler(url, depth, timeout)
	c.Crawl()
	domains := c.GetVisitedDomains()
	fmt.Println("Crawled domains: ")
	for _, domain := range domains {
		fmt.Println(domain)
	}
}
