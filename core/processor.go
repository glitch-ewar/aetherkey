package core

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

var queuedShodanRequests int32 = 0

func (s *Session) InitProcessors() {
	s.Processors = make(map[string]Processor)

	s.Processors["default"] = Processor{
		Columns: []string{"Repository", "Match"},
		Validate: func(signature, repository, match string) bool {
			return true
		},
	}

	s.Processors["Shodan API Key"] = Processor{
		Columns: []string{"Match"},
		Validate: func(signature, repository, match string) bool {
			//s.Log.Important("%s: %f", match, GetEntropy(match))

			//entropy := GetEntropy(match)

			atomic.AddInt32(&queuedShodanRequests, 1)
			time.Sleep(time.Second * time.Duration(atomic.LoadInt32(&queuedShodanRequests)))

			url := fmt.Sprintf("https://api.shodan.io/api-info?key=%s", match)
			resp, err := http.Get(url)
			// handle the error if there is one
			if err != nil {
				panic(err)
			}
			// do this now so it won't be forgotten
			defer resp.Body.Close()
			// reads html as a slice of bytes
			// html, err := ioutil.ReadAll(resp.Body)
			// if err != nil {
			// 	panic(err)
			// }

			// html = nil

			atomic.AddInt32(&queuedShodanRequests, -1)

			// show the HTML code as a string %s
			s.Log.Important("%s - %d", match, resp.StatusCode)

			return resp.StatusCode == 200
		},
	}

}

func (s *Session) GetProcessor(signature string) Processor {
	processor, contains := s.Processors[signature]
	if contains {
		return processor
	} else {
		return s.Processors["default"]
	}
}
