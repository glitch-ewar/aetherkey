package core

func (s *Session) InitProcessors() {
	s.Processors = make(map[string]Processor)

	s.Processors["default"] = Processor{Columns: []string{"Repository", "Match"}}

}

func (s *Session) GetProcessor(signature string) Processor {
	processor, contains := s.Processors[signature]
	if contains {
		return processor
	} else {
		return s.Processors["default"]
	}
}
