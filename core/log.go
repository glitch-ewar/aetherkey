package core

import (
	"fmt"
	"os"
	"regexp"
	"sync"

	"github.com/fatih/color"
)

const (
	FATAL     = 5
	ERROR     = 4
	IMPORTANT = 3
	WARN      = 2
	INFO      = 1
	DEBUG     = 0
)

var LogColors = map[int]*color.Color{
	FATAL:     color.New(color.FgRed).Add(color.Bold),
	ERROR:     color.New(color.FgRed),
	WARN:      color.New(color.FgYellow),
	IMPORTANT: color.New(),
	DEBUG:     color.New(color.Faint),
}

var firstEntry = true

type Logger struct {
	sync.Mutex

	debug  bool
	silent bool
}

func (l *Logger) SetDebug(d bool) {
	l.debug = d
}

func (l *Logger) SetSilent(d bool) {
	l.silent = d
}

func (l *Logger) Log(level int, format string, args ...interface{}) {
	l.Lock()
	defer l.Unlock()

	if level == DEBUG && !l.debug {
		return
	}

	if l.silent && level < IMPORTANT {
		return
	}

	if GetUI() == nil {
		if c, ok := LogColors[level]; ok {
			c.Printf("\r"+format+"\n", args...)
		} else {
			fmt.Printf("\r"+format+"\n", args...)
		}
	} else {
		text := colorStrip(fmt.Sprintf(format, args...))
		if firstEntry {
			fmt.Fprintf(GetUI().LogWindow, "%s", text)
			firstEntry = false
		} else {
			fmt.Fprintf(GetUI().LogWindow, "\r\n%s", text)
		}
	}

	if level == FATAL {
		os.Exit(1)
	}
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	l.Log(FATAL, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.Log(ERROR, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.Log(WARN, format, args...)
}

func (l *Logger) Important(format string, args ...interface{}) {
	l.Log(IMPORTANT, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.Log(INFO, format, args...)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.Log(DEBUG, format, args...)
}

func colorStrip(str string) string {
	ansi := "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	re := regexp.MustCompile(ansi)
	return re.ReplaceAllString(str, "")
}
