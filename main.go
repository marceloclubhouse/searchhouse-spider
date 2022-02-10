package main

import (
	"clubhouse/spider"
)

func main() {
	seed := []string{"https://marcelocubillos.com"}
	s := spider.ClubhouseSpider{}
	s.SetSeed(seed)
	s.SetWorkingDirectory("pages")
	s.Crawl()
	return
}
