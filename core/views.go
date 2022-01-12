package core

func (s *Session) InitViews() {
	s.Views = make(map[string][]string)

	s.Views["Default"] = []string{"Repository", "Match"}
	s.Views["Shodan API Key"] = []string{"Key", "Plan", "Query credits", "Scan credits"}
}

func (s *Session) GetView(signature string) []string {
	view, contains := s.Views[signature]
	if contains {
		return view
	} else {
		return s.Views["default"]
	}
}
