package postings

import (
	"strconv"
	"strings"
)

/*
	Posting object but with a sorted slice
	instead of a map.

    Sorted posting objects are used for creating
    and merging postings from the disk.

    Since the IDs are in order, a sorted posting
    uses a slice instead of a map to
    represent documentID IDs and respective
    frequencies or weights.

    This enables linear merging with relatively
    low memory footprints.
*/

type SortedPostingEntry struct {
	documentID uint64
	score      float64
}

type SortedPosting struct {
	token        string
	documents    []SortedPostingEntry
	docFrequency uint64
	length       uint64
}

func NewSortedPosting(token string, documents []SortedPostingEntry, docFrequency uint64, length uint64) SortedPosting {
	return SortedPosting{
		token:        token,
		documents:    documents,
		docFrequency: docFrequency,
		length:       length,
	}
}

func (s *SortedPosting) GetDocumentsAsList() []SortedPostingEntry {
	return s.documents
}

func (s *SortedPosting) GetLength() uint64 {
	return s.length
}

// SerializeEntries
//
// Return a string representing the documents
// and scores in this Posting.
func (s *SortedPosting) SerializeEntries() string {
	var documentsStr string
	for _, entry := range s.documents {
		documentsStr += strconv.FormatUint(entry.documentID, 10) + ":" + strconv.FormatFloat(entry.score, 'f', 2, 64) + ","
	}
	return strings.TrimSuffix(documentsStr, ",")
}

// DeserializeEntries
//
// Return a list of entries given a string
// representing the documents and scores of
// a Posting.
func DeserializeEntries(postingString string) []SortedPostingEntry {
	var entryList []SortedPostingEntry
	var docID string
	var score string
	pointerIsDocumentID := true
	lastCharIndex := len(postingString) - 1
	for charIndex, char := range postingString {
		if char == ':' {
			pointerIsDocumentID = false
		} else if char == ',' || charIndex == lastCharIndex {
			pointerIsDocumentID = true
			documentAsInt, err := strconv.ParseUint(docID, 10, 64)
			if err != nil {
				panic(err)
			}
			scoreAsFloat, err := strconv.ParseFloat(score, 64)
			if err != nil {
				panic(err)
			}
			entryList = append(entryList, SortedPostingEntry{documentID: documentAsInt, score: scoreAsFloat})
			docID = ""
			score = ""
		} else {
			if pointerIsDocumentID {
				docID += string(char)
			} else {
				score += string(char)
			}
		}
	}

	return entryList
}

// MergeEntries
//
// Take a 2D list of SortedPosting entries
// representing documentID ID and score
// and merge all of them into one list with
// growing linear time.
//
// Assume each documentID list is sorted.
func MergeEntries(postingEntries [][]SortedPostingEntry) []SortedPostingEntry {
	var mergedEntries []SortedPostingEntry
	// Construct cursors for each row of entries
	rowCursors := make([]int, len(postingEntries))
	entriesAtRowCursors := make([]*SortedPostingEntry, len(postingEntries))
	// Construct values at cursors
	for rowIndex := range postingEntries {
		rowCursors[rowIndex] = 0
		entriesAtRowCursors[rowIndex] = &postingEntries[rowIndex][rowCursors[rowIndex]]
	}
	for true {
		// If all the cursors have reached the end
		// of their respective rows, then terminate
		for rowIndex := range postingEntries {
			if rowCursors[rowIndex] != len(postingEntries[rowIndex]) {
				break
			}
			return mergedEntries
		}

		// Compute the smallest documentID ID out of
		// the entries selected on all cursors
		minDocumentID := ^uint64(0) // maximum uint64
		for _, entry := range entriesAtRowCursors {
			if entry != nil && entry.documentID < minDocumentID {
				minDocumentID = entry.documentID
			}
		}

		mergedEntry := SortedPostingEntry{
			documentID: minDocumentID,
			score:      0,
		}

		// Merge all entries with the smallest
		// document ID and append it to the slice
		for rowIndex := range rowCursors {
			if entriesAtRowCursors[rowIndex] != nil && entriesAtRowCursors[rowIndex].documentID == minDocumentID {
				mergedEntry.score += entriesAtRowCursors[rowIndex].score
				rowCursors[rowIndex] += 1
				if rowCursors[rowIndex] == len(postingEntries[rowIndex]) {
					entriesAtRowCursors[rowIndex] = nil
				} else {
					entriesAtRowCursors[rowIndex] = &postingEntries[rowIndex][rowCursors[rowIndex]]
				}
			}
		}

		mergedEntries = append(mergedEntries, mergedEntry)
	}

	// This is here so the compiler won't complain
	return mergedEntries
}
