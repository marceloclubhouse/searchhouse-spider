package indexer

import (
	"fmt"
	"regexp"
)

type Tokenizer struct {
}

func (t Tokenizer) FindAllTags(str string, tags []string) []string {
	// Extract the <p> tags out of HTML markup
	// and return the content of each tag as a
	// list of strings
	var content []string
	for _, tag := range tags {
		re := regexp.MustCompile(fmt.Sprintf(`<%s.*?>(.*)</%s>`, tag, tag))
		matches := re.FindAllStringSubmatch(str, -1)
		for _, element := range matches {
			content = append(content, element[1])
		}
	}
	return content
}

func (t Tokenizer) FindAllAnchors(str string) []string {
	// Find all of the links within HTML markup
	// (<a href="...">) -> ["..."]
	var content []string
	re := regexp.MustCompile(`href=[\'"]?([^\'" >]+)`)
	matches := re.FindAllStringSubmatch(str, -1)
	for _, element := range matches {
		content = append(content, element[1])
	}
	return content
}