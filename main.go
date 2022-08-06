package main

import (
	"clubhouse/spider"
)

func main() {
	numThreads := 100
	pageDir := "pages"
	seed := []string{"https://uci.edu/"}
	s := spider.New(numThreads, pageDir, seed)
	s.CrawlConcurrently()
	return
}
