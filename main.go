package main

import (
	"clubhouse/spider"
)

func main() {
	numThreads := 5
	pageDir := "pages"
	seed := []string{"https://marcelocubillos.com"}
	s := spider.New(numThreads, pageDir, seed)
	s.Crawl()
	return
}
