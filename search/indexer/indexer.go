package indexer

import (
	"bufio"
	"clubhouse/search/postings"
	"regexp"
	"strconv"
)

type Indexer struct {
	numRoutines   int
	indexLocation string
	verbose       bool
	weights       map[string]int
}

func NewIndexer(numRoutines int, indexLocation string, verbose bool, weights map[string]int) Indexer {
	return Indexer{
		numRoutines:   numRoutines,
		indexLocation: indexLocation,
		verbose:       verbose,
		weights:       weights,
	}
}

func readPostingFromDisk(inputStream *bufio.Reader) *SortedPosting {
	// Construct string from index file
	var postStr string
	// Read 3 lines at a time (format of a Posting)
	for i := 0; i < 3; i++ {
		newLine, err := inputStream.ReadString('\n')
		if err != nil {
			panic(err)
		}
		// If EOF is reached then return none
		if len(newLine) == 0 {
			return nil
		}
		postStr += newLine
	}

	// Read data from string and convert to
	// Golang objects
	re := regexp.MustCompile("^(?P<token>[a-z0-9]+)\n(?P<documents>(?:[0-9]+:[0-9.]+,)*(?:[0-9]+:[0-9.]+))\n(?P<doc_freq>[0-9.]+)\n?$")
	postingAttrs := re.FindStringSubmatch(postStr)
	if len(postingAttrs) == 0 {
		panic("Couldn't find match with " + postStr)
	}
	docs := postings.DeserializeEntries(postingAttrs[2])
	score, err := strconv.ParseFloat(postingAttrs[3], 64)
	if err != nil {
		panic(err)
	}
	sortedPosting := postings.NewSortedPosting(
		postingAttrs[1],
		docs,
		score,
		uint64(len(postStr)),
	)

	return &sortedPosting
}
