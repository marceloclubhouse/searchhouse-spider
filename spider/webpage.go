package spider

import (
	"encoding/json"
	"log"
)

type WebPage struct {
	Time     int64
	URL      string
	Response string
	Body     string
}

func (w WebPage) Serialize() []byte {
	// Serialize WebPage object as JSON byte array
	b, err := json.Marshal(w)
	if err != nil {
		log.Fatalln(err)
	}
	return b
}
