package spider

import (
	"clubhouse/indexer"
	"errors"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

type ClubhouseSpider struct {
	frontier         StringQueue
	tokenizer        indexer.Tokenizer
	workingDirectory string
}

func (s *ClubhouseSpider) Crawl() {
	for s.frontier.Len() > 0 {
		current_url := s.frontier.Pop()
		if !s.pageDownloaded(current_url) {
			resp, err := http.Get(current_url)
			if resp.Status == "200 OK" && err == nil {
				body, err := ioutil.ReadAll(resp.Body)
				if err == nil {
					fmt.Printf("<ClubhouseSpider.Crawl() - Response: %s, URL: %s>\n", resp.Status, current_url)
					page := WebPage{time.Now().Unix(), current_url, resp.Status, string(body)}
					s.writeToDisk(page)
					// Continue constructing frontier
					anchors := s.constructProperURLs(s.tokenizer.FindAllAnchors(string(body)), current_url)
					for key := range anchors.m {
						if !s.pageDownloaded(key) {
							s.frontier.Insert(key)
						}
					}
				}
			}
			time.Sleep(time.Second * 5)
		}
	}
}

func (s *ClubhouseSpider) writeToDisk(w WebPage) {
	// Serialize web page to JSON format
	fileName := s.workingDirectory + "/" + strconv.FormatUint(s.hash(w.URL), 10) + ".json"
	filePath, err := filepath.Abs(fileName)
	if err != nil {
		log.Fatalln(err)
	}
	f, err := os.Create(filePath)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = f.Write(w.Serialize())
	if err != nil {
		log.Fatalln(err)
	}
	err = f.Close()
	if err != nil {
		log.Fatalln(err)
	}
}

func (s *ClubhouseSpider) fileExists(path string) (bool, error) {
	// https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func (s ClubhouseSpider) pageDownloaded(url string) bool {
	fileName := s.workingDirectory + "/" + strconv.FormatUint(s.hash(url), 10) + ".json"
	exists, err := s.fileExists(fileName)
	if err != nil {
		log.Fatalln(err)
	}
	if exists {
		return true
	} else {
		return false
	}
}

func (s *ClubhouseSpider) urlValid(url string) bool {
	// Return True if a URL is valid, False otherwise
	// URL must not have fragment (#) and not end
	// with a non-HTML file extension
	re := regexp.MustCompile(`https://(www\.)?[-a-zA-Z0-9@:%._+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_+.~#?&/=]*)`)
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

func (s *ClubhouseSpider) hash(str string) uint64 {
	h := fnv.New64a()
	_, err := h.Write([]byte(str))
	if err != nil {
		log.Fatalln(err)
	}
	return h.Sum64()
}

func (s *ClubhouseSpider) SetSeed(urls []string) {
	for _, url := range urls {
		s.frontier.Insert(url)
	}
}

func (s *ClubhouseSpider) SetWorkingDirectory(dir string) {
	s.workingDirectory = dir
}

func (s ClubhouseSpider) PrintFrontier() {
	fmt.Println(s.frontier)
}
