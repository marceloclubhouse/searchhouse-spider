package main

import (
	"clubhouse/spider"
)

func main() {
	numRoutines := 1000
	maxLinks := 100
	pageDir := "/Users/marcelo/clubhouse-pages"
	seed := []string{"https://marcelocubillos.com"}
	s := spider.New(numRoutines, pageDir, seed, maxLinks)
	s.CrawlConcurrently()
	return
}
