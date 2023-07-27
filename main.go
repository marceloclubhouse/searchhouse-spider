package main

import (
	"clubhouse-spider/spider"
	"flag"
)

func main() {
	// Frontier (pages.db) must be reset if numRoutines changes in between runs!
	numRoutines := flag.Int("numRoutines", 1, "Number of routines for spider to use")
	pageDir := flag.String("pageDir", "pages", "Location for pages to be saved")
	seed := flag.String("seed", "", "First page to start out crawling with")
	flag.Parse()
	s := spider.New(*numRoutines, *pageDir, []string{*seed})
	s.Crawl()
	return
}
