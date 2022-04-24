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
	"strings"
	"time"
)

type ClubhouseSpider struct {
	numThreads       int
	frontier         Frontier
	pages            indexer.Pages
	tokenizer        indexer.Tokenizer
	workingDirectory string
}

func New(numThreads int, workingDirectory string, seed []string) ClubhouseSpider {
	cs := ClubhouseSpider{numThreads, Frontier{}, indexer.Pages{}, indexer.Tokenizer{}, workingDirectory}
	cs.frontier.Init()
	cs.pages.Init()
	cs.setSeed(seed)
	return cs
}

func (s *ClubhouseSpider) Crawl() {
	for true {
		currentUrl := s.frontier.PopURL()
		if currentUrl == "" {
			return
		}
		if !s.pageDownloaded(currentUrl) {
			resp, err := http.Get(currentUrl)
			if err == nil {
				fmt.Printf("<ClubhouseSpider.Crawl() - Response: %s, URL: %s>\n", resp.Status, currentUrl)
				if resp.Status == "200 OK" {
					body, err := ioutil.ReadAll(resp.Body)
					s.pages.InsertPage(currentUrl)
					if err == nil {
						page := WebPage{time.Now().Unix(), currentUrl, resp.Status, string(body)}
						s.writeToDisk(page)
						// Continue constructing frontier
						anchors := s.constructProperURLs(s.tokenizer.FindAllAnchors(string(body)), currentUrl)
						for key := range anchors.m {
							if !s.pageDownloaded(key) && !s.frontier.CheckURLInFrontier(key) {
								s.frontier.InsertPage(key)
							}
						}
					}
				}
			}
			time.Sleep(time.Second)
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
	urlRe := regexp.MustCompile(`^https://(www\.)?[-a-zA-Z0-9@:%._+~#=]{1,256}\.[a-zA-Z0-9()]` +
		`{1,6}\b([-a-zA-Z0-9()@:_+~/]*)$`)
	extRe := regexp.MustCompile(`.*\.(?:css|js|bmp|gif|jpe?g|ico|png|tiff?|mid|mp2|mp3|mp4|ppsx|` +
		`wav|avi|mov|mpeg|ram|m4v|mkv|ogg|ogv|pdf|odc|sas|ps|eps|tex|ppt|pptx|doc|docx|xls|xlsx|` +
		`names|data|dat|exe|bz2|tar|msi|bin|7z|psd|dmg|iso|epub|dll|cnf|tgz|sha1|ss|scm|py|rkt|r|c|` +
		`thmx|mso|arff|rtf|jar|csv|java|txt|rm|smil|wmv|swf|wma|zip|rar|gz)$`)
	if urlRe.MatchString(url) && !extRe.MatchString(strings.ToLower(url)) {
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
	domainRe := regexp.MustCompile(`https?://[^\s:/]+\.[^\s:/]+`)
	root = domainRe.FindAllStringSubmatch(root, -1)[0][0]

	for _, url := range urls {
		var parsedURL string
		if url[0] == '/' {
			parsedURL = root + url
		} else {
			if url[len(url)-1] == '/' {
				parsedURL = url[:len(url)-1]
			} else {
				parsedURL = url
			}
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

func (s *ClubhouseSpider) setSeed(urls []string) {
	for _, url := range urls {
		s.frontier.InsertPage(url)
	}
}

func (s ClubhouseSpider) PrintFrontier() {
	fmt.Println(s.frontier)
}
