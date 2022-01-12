package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var shodanKeyCache map[string]bool
var shodanNextAPICall time.Time

type ShodanAPIInfo struct {
	ScanCredits  int    `json:"scan_credits"`
	QueryCredits int    `json:"query_credits"`
	Plan         string `json:"plan"`
}

// {
//     "scan_credits": 100000,
//     "usage_limits": {
//         "scan_credits": -1,
//         "query_credits": -1,
//         "monitored_ips": -1
//     },
//     "plan": "stream-100",
//     "https": false,
//     "unlocked": true,
//     "query_credits": 100000,
//     "monitored_ips": 19,
//     "unlocked_left": 100000,
//     "telnet": false
// }

func (s *Session) InitValidators() {
	s.Validators = make(map[string]Validator)

	s.Validators["default"] = func(signature string, match string) (bool, ValidationInfo, Relevance) {
		return true, ValidationInfo{}, RelevanceMedium
	}

	shodanKeyCache = map[string]bool{}
	shodanNextAPICall = time.Now()
	s.Validators["Shodan API Key"] = func(signature string, match string) (bool, ValidationInfo, Relevance) {
		info := ValidationInfo{}
		relevance := RelevanceLow

		exists, _ := shodanKeyCache[match]
		if exists {
			return false, info, relevance
		} else {

			if time.Now().Before(shodanNextAPICall) {
				time.Sleep(time.Until(shodanNextAPICall))
			}
			shodanNextAPICall = time.Now().Add(time.Second * 2)

			url := fmt.Sprintf("https://api.shodan.io/api-info?key=%s", match)
			resp, err := http.Get(url)
			if err != nil {
				s.Log.Error("Error while retrieving %s: %s", url, err.Error())
				return false, info, relevance
			}
			defer resp.Body.Close()

			isValidKey := resp.StatusCode == 200
			if resp.StatusCode != 200 && resp.StatusCode != 401 {
				s.Log.Important("Shodan validation error: %d.", resp.StatusCode)
			} else if isValidKey {
				rawData, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					s.Log.Error("Error while reading HTML data %s: %s", url, err.Error())
					return false, info, relevance
				}

				apiInfo := ShodanAPIInfo{}
				json.Unmarshal(rawData, &apiInfo)

				info["Key"] = match
				info["Plan"] = apiInfo.Plan
				info["Query credits"] = fmt.Sprint(apiInfo.QueryCredits)
				info["Scan credits"] = fmt.Sprint(apiInfo.ScanCredits)

				if apiInfo.QueryCredits > 0 || apiInfo.ScanCredits > 0 {
					relevance = RelevanceHigh
				}
			}

			shodanKeyCache[match] = true
			return isValidKey, info, relevance
		}
	}
}

func (s *Session) GetValidator(signature string) Validator {
	validator, contains := s.Validators[signature]
	if contains {
		return validator
	} else {
		return s.Validators["default"]
	}
}
