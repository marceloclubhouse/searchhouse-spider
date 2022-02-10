package main

import (
	"clubhouse/spider"
)

func main() {
	seed := []string{"https://marcelocubillos.com"}
	spider := spider.ClubhouseSpider{}
	spider.SetSeed(seed)
	spider.Crawl()
	return
}
