package main

import (
	"clubhouse/spider"
)

func main() {
	numRoutines := 100
	maxLinks := 100
	pageDir := "pages"
	seed := []string{"https://uci.edu/"}
	s := spider.New(numRoutines, pageDir, seed, maxLinks)
	s.CrawlConcurrently()
	return
}
