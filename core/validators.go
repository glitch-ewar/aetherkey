package core

import (
	"fmt"
	"net/http"
	"time"
)

var shodanKeyCache map[string]bool
var shodanNextAPICall time.Time

func (s *Session) InitValidators() {
	s.Validators = make(map[string]Validator)

	s.Validators["default"] = func(signature string, match string) (bool, ValidationInfo) {
		return true, ValidationInfo{}
	}

	shodanKeyCache = map[string]bool{}
	shodanNextAPICall = time.Now()
	s.Validators["Shodan API Key"] = func(signature string, match string) (bool, ValidationInfo) {
		info := ValidationInfo{}

		exists, _ := shodanKeyCache[match]
		if exists {
			return false, info
		} else {

			if time.Now().Before(shodanNextAPICall) {
				time.Sleep(time.Until(shodanNextAPICall))
			}
			shodanNextAPICall = time.Now().Add(time.Second * 2)

			url := fmt.Sprintf("https://api.shodan.io/api-info?key=%s", match)
			resp, err := http.Get(url)
			if err != nil {
				s.Log.Error("Error while retrieving %s: %s", url, err.Error())
				return false, info
			}
			defer resp.Body.Close()
			// reads html as a slice of bytes
			// html, err := ioutil.ReadAll(resp.Body)
			// if err != nil {
			// 	panic(err)
			// }

			isValidKey := resp.StatusCode == 200
			if resp.StatusCode != 200 && resp.StatusCode != 401 {
				s.Log.Important("Shodan validation error: %d.", resp.StatusCode)
			}

			shodanKeyCache[match] = true
			return isValidKey, info
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
