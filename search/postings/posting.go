package postings

import (
	"sort"
	"strconv"
	"strings"
)

/*
	Posting object is used to represent a term
	within an index.

	token           - alphanumeric sequence
    docFrequency    - number of docs the term appears in
    documents       - dictionary of documents and their
                      respective frequencies

    This Posting object is designed to be initialized
    with a token, and to have documents repeatedly
    added to it. This is different from
    postings.sorted_posting.
*/

type Posting struct {
	token        string
	docFrequency uint64
	documents    map[uint64]uint64
}

func NewPosting(token string) Posting {
	return Posting{
		token,
		0,
		make(map[uint64]uint64),
	}
}

func (p *Posting) GetToken() string {
	return p.token
}

func (p *Posting) GetDocFrequency() uint64 {
	return p.docFrequency
}

func (p *Posting) GetDocuments() map[uint64]uint64 {
	return p.documents
}

// AddOccurrence
// Insert a documentID into this Posting.
//
// If this documentID was already inserted before,
// increment the frequency of this token in the
// documentID.
//
// If this documentID was not already inserted,
// also increment the number of documents this
// token appears in.
func (p *Posting) AddOccurrence(id uint64, freq uint64) {
	_, exists := p.documents[id]
	// If the documentID already exists in this posting
	if exists {
		p.documents[id] += freq
	} else {
		p.documents[id] = freq
		p.docFrequency += 1
	}
}

// SerializeToSortedPostingString
//
// Return string of sorted documentID ids and
// term frequencies as a string of id:freq
//
// (e.g 0:5,2:19 - term occurs 5 times
// in doc 0, and 19 times in doc 2)
func (p *Posting) SerializeToSortedPostingString() string {
	var documentsStr string
	keys := make([]uint64, 0, len(p.documents))
	for key := range p.documents {
		keys = append(keys, key)
	}
	sort.Sort(uint64Slice(keys))
	for _, key := range keys {
		documentsStr += strconv.FormatUint(key, 10) + ":" + strconv.FormatUint(p.documents[key], 10) + ","
	}
	return strings.TrimSuffix(documentsStr, ",")
}

// A custom type that implements sort.Interface for uint64 values
type uint64Slice []uint64

func (s uint64Slice) Len() int {
	return len(s)
}

func (s uint64Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s uint64Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
