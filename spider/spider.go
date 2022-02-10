package spider

import (
	"clubhouse/indexer"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

type ClubhouseSpider struct {
	frontier StringQueue
	tokenizer indexer.Tokenizer
}

func (s *ClubhouseSpider) Crawl() {
	for s.frontier.Len() > 0 {
		current_url := s.frontier.Pop()
		resp, err := http.Get(current_url)
		if err == nil {
			body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				fmt.Printf("<ClubhouseSpider.Crawl() - Response: %s, URL: %s>\n", resp.Status, current_url)
				content := string(body)
				anchors := s.constructProperURLs(s.tokenizer.FindAllAnchors(content), current_url)
				for key,_ := range anchors.m {
					s.frontier.Insert(key)
				}
			}
		}
		time.Sleep(time.Second)
	}
}

func (s *ClubhouseSpider) serializeContent(url string, body string) {
	// Serialize web page to disk in JSON format
	// TODO
	return
}

func (s *ClubhouseSpider) urlValid(url string) bool {
	// Return True if a URL is valid, False otherwise
	// URL must not have fragment (#) and not end
	// with a non-HTML file extension
	re := regexp.MustCompile(`https:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)
	if re.MatchString(url) {
		return true
	} else {
		return false
	}
}

func (s *ClubhouseSpider) constructProperURLs(urls []string, root string) StringSet {
	// Given a list of anchor HREF strings and a
	// root URL, return a list of proper URLS
	// (i.e. /clubhouse -> https://mc.com/clubhouse)

	// Remove all fragments (#), duplicate URLs, and
	// return unordered list of URLs as StringSet

	var properURLs StringSet
	if root[len(root)-1] == '/' {
		root = root[:len(root)-1]
	}

	for _, url := range urls {
		var parsedURL string
		if url[0] == '/' {
			parsedURL = root + url
		} else {
			parsedURL = url
		}
		if s.urlValid(parsedURL) {
			properURLs.Add(parsedURL)
		}
	}

	return properURLs
}

func (s *ClubhouseSpider) SetSeed(urls []string) {
	for _, url := range urls {
		s.frontier.Insert(url)
	}
}

func (s ClubhouseSpider) PrintFrontier() {
	fmt.Println(s.frontier)
}