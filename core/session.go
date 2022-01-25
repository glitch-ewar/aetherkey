package core

import (
	"context"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Columns []string

type SearchResult struct {
	Signature Signature
	Url       string
	RawUrl    string
}

type ValidationInfo map[string]string
type Validator func(signature string, match string) (bool, ValidationInfo, Relevance)
type CsvWriters map[string]*csv.Writer

type Session struct {
	sync.Mutex

	Version          string
	Log              *Logger
	Options          *Options
	Config           *Config
	Signatures       []Signature
	Repositories     chan GitResource
	Gists            chan string
	Comments         chan string
	SearchResults    chan SearchResult
	Context          context.Context
	Clients          chan *GitHubClientWrapper
	ExhaustedClients chan *GitHubClientWrapper
	Views            map[string][]string
	Validators       map[string]Validator
	CsvWriters       CsvWriters
}

var (
	session     *Session
	sessionSync sync.Once
	err         error
)

func (s *Session) Start() {
	rand.Seed(time.Now().Unix())

	s.InitLogger()
	s.InitViews()
	s.InitValidators()
	s.InitThreads()
	s.InitSignatures()
	s.InitGitHubClients()
	s.InitCsvWriters()
}

func (s *Session) InitLogger() {
	s.Log = &Logger{}
	s.Log.SetDebug(*s.Options.Debug)
	s.Log.SetSilent(*s.Options.Silent)
}

func (s *Session) InitSignatures() {
	s.Signatures = GetSignatures(s)
}

func (s *Session) InitGitHubClients() {
	if len(*s.Options.Local) <= 0 {
		chanSize := *s.Options.Threads * (len(s.Config.GitHubAccessTokens) + 1)
		s.Clients = make(chan *GitHubClientWrapper, chanSize)
		s.ExhaustedClients = make(chan *GitHubClientWrapper, chanSize)
		for _, token := range s.Config.GitHubAccessTokens {
			ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
			tc := oauth2.NewClient(s.Context, ts)

			client := github.NewClient(tc)
			client.UserAgent = fmt.Sprintf("%s v%s", Name, Version)
			_, _, err := client.Users.Get(s.Context, "")

			if err != nil {
				if _, ok := err.(*github.ErrorResponse); ok {
					s.Log.Warn("Failed to validate token %s[..]: %s", token[:10], err)
					continue
				}
			}

			for i := 0; i <= *s.Options.Threads; i++ {
				s.Clients <- &GitHubClientWrapper{client, token, time.Now().Add(-1 * time.Second)}
			}
		}

		if len(s.Clients) < 1 {
			s.Log.Fatal("No valid GitHub tokens provided. Quitting!")
		}
	}
}

func (s *Session) GetClient() *GitHubClientWrapper {
	for {
		select {

		case client := <-s.Clients:
			s.Log.Debug("Using client with token: %s", client.Token[:10])
			return client

		case client := <-s.ExhaustedClients:
			sleepTime := time.Until(client.RateLimitedUntil)
			s.Log.Warn("All GitHub tokens exhausted/rate limited. Sleeping for %s", sleepTime.String())
			time.Sleep(sleepTime)
			s.Log.Debug("Returning client %s to pool", client.Token[:10])
			s.FreeClient(client)

		default:
			s.Log.Debug("Available Clients: %d", len(s.Clients))
			s.Log.Debug("Exhausted Clients: %d", len(s.ExhaustedClients))
			time.Sleep(time.Millisecond * 1000)
		}
	}
}

// FreeClient returns the GitHub Client to the pool of available,
// non-rate-limited channel of clients in the session
func (s *Session) FreeClient(client *GitHubClientWrapper) {
	if client.RateLimitedUntil.After(time.Now()) {
		s.ExhaustedClients <- client
	} else {
		s.Clients <- client
	}
}

func (s *Session) InitThreads() {
	if *s.Options.Threads == 0 {
		numCPUs := runtime.NumCPU()
		s.Options.Threads = &numCPUs
	}

	runtime.GOMAXPROCS(*s.Options.Threads + 1)
}

func (s *Session) getCsvDir() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		fmt.Println("Unable to get cache directory:", err)
		os.Exit(1)
	}

	csvDir := fmt.Sprintf("%s%caetherkey", cacheDir, os.PathSeparator)
	os.Mkdir(csvDir, 0755)

	return csvDir
}

func (s *Session) InitCsvWriters() {
	s.CsvWriters = make(CsvWriters)
	csvDir := s.getCsvDir()
	for _, signature := range s.Signatures {
		csvPath := fmt.Sprintf("%s%c%s.csv", csvDir, os.PathSeparator, signature.Name())

		writeHeader := false
		if !PathExists(csvPath) {
			writeHeader = true
		}

		file, err := os.OpenFile(csvPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Could not create CSV file:", err)
			continue
		} else {
			writer := csv.NewWriter(file)
			s.CsvWriters[signature.Name()] = writer
			if writeHeader {
				header := []string{"File", "Match", "URL"}
				header = append(header, s.Views[signature.Name()]...)
				writer.Write(header)
				writer.Flush()
			}

		}
	}
}

func (s *Session) WriteToCsv(event *MatchEvent) {
	writer, exists := s.CsvWriters[event.Signature]
	if exists == false {
		return
	}

	line := []string{
		event.File,
		event.Match,
		event.Url,
	}

	for _, v := range event.AdditionalInfo {
		line = append(line, v)
	}

	writer.Write(line)
	writer.Flush()
}

func (s *Session) LoadCsvs() {
	csvDir := s.getCsvDir()
	for _, signature := range s.Signatures {
		csvPath := fmt.Sprintf("%s%cAetherKey%c%s.csv", csvDir, os.PathSeparator, os.PathSeparator, signature.Name())
		file, err := os.Open(csvPath)
		if err != nil {
			s.Log.Error("Could not open CSV file: %s.", err)
			continue
		} else {
			defer file.Close()
			reader := csv.NewReader(file)
			reader.ReadAll()
		}
	}
}

func GetSession() *Session {
	sessionSync.Do(func() {
		session = &Session{
			Context:       context.Background(),
			Repositories:  make(chan GitResource, 1000),
			Gists:         make(chan string, 100),
			Comments:      make(chan string, 1000),
			SearchResults: make(chan SearchResult, 1000),
		}

		if session.Options, err = ParseOptions(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if session.Config, err = ParseConfig(session.Options); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		session.Start()
	})

	return session
}
