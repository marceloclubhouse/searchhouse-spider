package spider

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"regexp"
	"strings"
)

type WebPage struct {
	Time     int64
	URL      string
	Response string
	Body     string
}

func (w *WebPage) Serialize() []byte {
	// Serialize WebPage object as JSON byte array
	b, err := json.Marshal(w)
	if err != nil {
		log.Fatalln(err)
	}
	return b
}

func (w *WebPage) FindAllAnchorHREFs(maxNumHREF int) []string {
	// Find all of the links within HTML markup
	// (<a href="...">) -> ["..."]
	var hrefs []string
	re := regexp.MustCompile(`href=['"]?([^'" >]+)`)
	matches := re.FindAllStringSubmatch(w.Body, maxNumHREF)
	for _, element := range matches {
		hrefs = append(hrefs, element[1])
	}
	return hrefs
}

func (w *WebPage) findAllTags(tags []string) []string {
	// Extract the specified tags out of HTML
	// markup and return the content of each
	// tag as a list of strings
	var content []string
	for _, tag := range tags {
		re := regexp.MustCompile(fmt.Sprintf(`<%s.*?>(.*)</%s>`, tag, tag))
		matches := re.FindAllStringSubmatch(w.Body, -1)
		for _, element := range matches {
			content = append(content, element[1])
		}
	}
	return content
}

func (w *WebPage) removeAllMarkup(str string) string {
	// Given some text, remove all HTML brackets
	// from it (e.g. "this <span>is</span> text"
	// becomes "this is text")
	str = html.UnescapeString(str)
	re := regexp.MustCompile(`<.*?>|</.*>`)
	if re.MatchString(str) {
		return re.ReplaceAllString(str, " ")
	} else {
		return str
	}
}

func (w *WebPage) removeAllPunctuation(str string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9-:\s'’/–-]`)
	if re.MatchString(str) {
		return re.ReplaceAllString(str, " ")
	} else {
		return str
	}
}

func (w *WebPage) extractText() string {
	// Extract all the text from this WebPage object
	var str string
	tags := w.findAllTags([]string{"p"})
	for _, content := range tags {
		str = str + strings.ToLower(w.removeAllPunctuation(w.removeAllMarkup(content)))
	}
	return str
}

func (w *WebPage) generateNGrams(n int) string {
	return w.extractText()
}
