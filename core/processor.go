package core

var queuedShodanRequests int32 = 0

func (s *Session) InitProcessors() {
	s.Processors = make(map[string]Processor)

	s.Processors["default"] = Processor{
		Columns: []string{"Repository", "Match"},
		Validate: func(signature, repository, match string) bool {
			return true
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
