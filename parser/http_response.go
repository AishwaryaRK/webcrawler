package parser

import (
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type HttpClient struct {
	client *http.Client
}

func NewHttpClient(timeout int) *HttpClient {
	return &HttpClient{client: &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	},
	}
}

func (hc *HttpClient) FindUrlsOnPage(url string) []string {
	htmlDoc := hc.getParsedHtmlResponse(url)
	return hc.traverseHtmlDoc(htmlDoc)
}

func (hc *HttpClient) traverseHtmlDoc(node *html.Node) []string {
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
		childNodeUrls := hc.traverseHtmlDoc(childNode)
		urls = append(urls, childNodeUrls...)
	}

	return urls
}

func (hc *HttpClient) getParsedHtmlResponse(url string) *html.Node {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// fmt.Errorf("Error creating new http request. Error: ", err)
		log.Printf("Error creating new http request. Error: %v", err)
		return nil
	}

	resp, err := hc.client.Do(req)
	if err != nil {
		log.Printf("Error fetching http request's response. Error: %v", err)
		return nil
	}

	defer resp.Body.Close()

	htmlDoc, err := html.Parse(resp.Body)
	if err != nil {
		log.Printf("Error parsing html response body. Error: %v", err)
		return nil
	}
	return htmlDoc
}
