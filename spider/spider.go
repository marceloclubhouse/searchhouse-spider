package spider

import (
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

type SearchHouseSpider struct {
	numRoutines      int
	frontier         Frontier
	workingDirectory string
	maxLinksPerPage  int
	ioMu             *sync.Mutex
}

func NewSpider(numRoutines int, workingDirectory string, seed []string, maxLinks int) *SearchHouseSpider {
	ioMu := new(sync.Mutex)
	cs := SearchHouseSpider{numRoutines, Frontier{}, workingDirectory, maxLinks, ioMu}
	cs.frontier.Init()
	cs.setSeed(seed, ioMu)
	return &cs
}

func (s *SearchHouseSpider) CrawlConcurrently() {
	wg := new(sync.WaitGroup)
	wg.Add(s.numRoutines)
	for i := 0; i < s.numRoutines; i++ {
		go s.Crawl(i, wg, s.ioMu)
	}
	wg.Wait()
}

func (s *SearchHouseSpider) Crawl(routineNum int, wg *sync.WaitGroup, ioMu *sync.Mutex) {
	defer wg.Done()
	fp := NewFingerprints(3, 10000)
	for true {
		currentUrl := s.frontier.PopURL(routineNum)
		if currentUrl == "" || !s.urlValid(currentUrl) {
			time.Sleep(time.Second)
			continue
		}
		if !s.pageDownloaded(currentUrl, ioMu) {
			resp, err := http.Get(currentUrl)
			if err == nil {
				fmt.Printf("<SearchHouseSpider.Crawl(%d) - Response: %s, URL: %s>\n", routineNum, resp.Status, currentUrl)
				if resp.Status == "200 OK" {
					body, err := io.ReadAll(resp.Body)
					if err == nil {
						page := NewWebPage(time.Now().Unix(), currentUrl, resp.Status, string(body))
						err = resp.Body.Close()
						if err != nil {
							panic(err)
						}
						// Check for issues with the page before cataloging
						if !s.validPage(page) {
							fmt.Printf("<SearchHouseSpider.Crawl(%d) - Skipped %s since the HTML does not appear valid\n", routineNum, currentUrl)
							continue
						}
						if s.duplicateExists(fp, page) {
							fmt.Printf("<SearchHouseSpider.Crawl(%d) - Skipped %s since it has a near match\n", routineNum, currentUrl)
							continue
						}
						fp.InsertFingerprintsUsingWebpage(page)
						s.writeToDisk(*page, ioMu)
						// Continue constructing frontier
						anchors := s.constructProperURLs(page.FindAllAnchorHREFs(s.maxLinksPerPage), currentUrl)
						for key := range anchors.m {
							if !s.pageDownloaded(key, ioMu) {
								s.frontier.InsertPage(key, s.calcWebsiteToRoutineNum(key))
							}
						}
					} else {
						err = resp.Body.Close()
						if err != nil {
							panic(err)
						}
					}
				}
			}
			time.Sleep(time.Second)
		}
	}
}

func (s *SearchHouseSpider) writeToDisk(w WebPage, ioMu *sync.Mutex) {
	// Serialize web page to JSON format
	fileName := s.workingDirectory + "/" + strconv.FormatUint(s.hash(w.Url), 10) + ".json"
	defer ioMu.Unlock()
	ioMu.Lock()
	filePath, err := filepath.Abs(fileName)
	if err != nil {
		log.Fatalln(err)
	}
	f, err := os.Create(filePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(f)
	_, err = f.Write(w.Serialize())
	if err != nil {
		log.Fatalln(err)
	}
}

func (s *SearchHouseSpider) fileExists(path string) (bool, error) {
	// https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func (s *SearchHouseSpider) pageDownloaded(url string, ioMu *sync.Mutex) bool {
	// Check if a page is downloaded by generating
	// the corresponding filename and "pinging"
	// the filesystem
	fileName := s.workingDirectory + "/" + strconv.FormatUint(s.hash(url), 10) + ".json"
	defer ioMu.Unlock()
	ioMu.Lock()
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

func (s *SearchHouseSpider) urlValid(url string) bool {
	// Return True if a URL is valid, False otherwise
	// URL must not have fragment (#) and not end
	// with a non-HTML file extension
	urlRe := regexp.MustCompile(`^(?P<scheme>https://)` +
		`(?P<hostname>[-a-zA-Z0-9@:%._+~#=]{1,256}\.[a-zA-Z0-9()]{1,6})` +
		`(?P<resource>[-a-zA-Z0-9()@:_+~?=/]*)$`)
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

func (s *SearchHouseSpider) regexToMap(r *regexp.Regexp, str string) map[string]string {
	// Convert regex capturing groups and their
	// corresponding matches to a hashmap
	match := r.FindStringSubmatch(str)
	results := map[string]string{}
	for i, name := range match {
		results[r.SubexpNames()[i]] = name
	}
	return results
}

func (s *SearchHouseSpider) findHostName(url string) string {
	// Return the host name from a URL
	domainRe := regexp.MustCompile(`https://[^\s:/@]+\.[^\s:/@]+`)
	substr := domainRe.FindAllStringSubmatch(url, -1)
	if len(substr) == 0 || len(substr[0]) == 0 {
		return ""
	}
	return substr[0][0]
}

func (s *SearchHouseSpider) constructProperURLs(urls []string, root string) StringSet {
	// Given a list of anchor HREF strings and a
	// root URL, return a list of proper URLS
	// (i.e. /clubhouse -> https://mc.com/clubhouse)

	// Remove all fragments (#), duplicate URLs, and
	// return unordered list of URLs as StringSet
	var properURLs StringSet

	hostName := s.findHostName(root)
	// If the root isn't a proper URL, return an empty
	// list (can happen if link is mailto:)
	if hostName == "" {
		return properURLs
	}

	for _, url := range urls {
		var parsedURL string
		if url[0] == '/' {
			parsedURL = hostName + url
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

func (s *SearchHouseSpider) hash(str string) uint64 {
	// Given a string, return an unsigned int hash
	h := fnv.New64a()
	_, err := h.Write([]byte(str))
	if err != nil {
		log.Fatalln(err)
	}
	return h.Sum64()
}

func (s *SearchHouseSpider) calcWebsiteToRoutineNum(url string) int {
	hostname := s.findHostName(url)
	hash := s.abs(int(s.hash(hostname)))
	return hash % s.numRoutines
}

func (s *SearchHouseSpider) abs(val int) int {
	if val < 0 {
		return -1 * val
	} else {
		return val
	}
}

func (s *SearchHouseSpider) setSeed(urls []string, ioMu *sync.Mutex) {
	// Set the spider's seed by inserting URLs
	// into the frontier
	for _, url := range urls {
		if !s.pageDownloaded(url, ioMu) {
			s.frontier.InsertPage(url, 0)
		}
	}
}

func (s *SearchHouseSpider) duplicateExists(fp *Fingerprints, wp *WebPage) bool {
	fpGlobalSet := fp.GetFingerprintsAsSet()
	fpWebpageSet := wp.fingerprints.GetFingerprintsAsSet()
	fp.mu.Lock()
	wp.fingerprints.mu.Lock()
	defer wp.fingerprints.mu.Unlock()
	defer fp.mu.Unlock()
	for hash := range fpWebpageSet {
		if _, exists := fpGlobalSet[hash]; exists {
			for page := range fpGlobalSet[hash] {
				if page.Url != wp.Url {
					similarity := wp.Similarity(page)
					if similarity > 0.9 {
						fmt.Printf("%s has a %f match to %s\n", page.Url, similarity, wp.Url)
						return true
					}
				}
			}
		}
	}
	return false
}

func (s *SearchHouseSpider) validPage(wp *WebPage) bool {
	// Could eventually be expanded into evaluating the page in its
	// entirety to be valid HTML
	trimmedBody := strings.TrimLeftFunc(wp.Body, unicode.IsSpace)
	if strings.HasPrefix(trimmedBody, "<!DOCTYPE html") {
		return true
	} else if strings.HasPrefix(trimmedBody, "<!doctype html") {
		return true
	}
	return false
}
