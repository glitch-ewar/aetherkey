package core

import (
	"container/list"
	"fmt"
	"time"

	"github.com/rivo/tview"
)

type UI struct {
	App              *tview.Application
	MainWindow       *tview.Flex
	StatusWindow     *tview.TextView
	SignaturesWindow *tview.List
	LogWindow        *tview.TextView
}

type Result struct {
	Key string
}

var tui UI
var signatures map[string]*list.List

func GetUI() *UI {
	return &tui
}

func (ui *UI) Initialize() {
	signatures = make(map[string]*list.List)
	signatures["test"] = list.New()

	ui.App = tview.NewApplication()

	ui.StatusWindow = tview.NewTextView()
	ui.StatusWindow.SetBorder(true)
	ui.StatusWindow.SetTitle("AETHERKEY")
	ui.StatusWindow.SetChangedFunc(func() {
		ui.App.Draw()
	})

	ui.LogWindow = tview.NewTextView()
	ui.LogWindow.SetBorder(true)
	ui.LogWindow.SetTitle("Log")
	ui.LogWindow.SetChangedFunc(func() {
		ui.App.Draw()
	})

	ui.SignaturesWindow = tview.NewList()
	ui.SignaturesWindow.SetBorder(true)
	ui.SignaturesWindow.SetTitle("Signatures")

	hflex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(ui.SignaturesWindow, 0, 1, false).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Details"), 0, 6, false)

	ui.MainWindow = tview.NewFlex().SetDirection(tview.FlexRow)
	ui.MainWindow.AddItem(ui.StatusWindow, 3, 1, false)
	ui.MainWindow.AddItem(hflex, 0, 1, false)
	ui.MainWindow.AddItem(ui.LogWindow, 8, 1, false)

	go ui.UpdateStatus()
}

func (ui *UI) Publish(signature string) {
	results, contains := signatures[signature]
	if contains {
		indices := ui.SignaturesWindow.FindItems(signature, "", false, false)
		if len(indices) > 0 {

			result := Result{Key: "test"}
			results.PushBack(result)

			ui.SignaturesWindow.SetItemText(indices[0], signature, fmt.Sprintf("> %d entries", results.Len()))
		}
	} else {
		ui.SignaturesWindow.AddItem(signature, "> 1 entry", 0, nil)

		results := list.New()
		result := Result{Key: "test"}
		results.PushBack(result)
		signatures[signature] = results
	}

	//ui.SignaturesWindow.FindItems(signature, "", false, false)
}

func (ui *UI) Run() {
	if err := ui.App.SetRoot(ui.MainWindow, true).SetFocus(ui.SignaturesWindow).Run(); err != nil {
		panic(err)
	}
}

func (ui *UI) UpdateStatus() {
	const refreshInterval = 500 * time.Millisecond
	for {
		time.Sleep(refreshInterval)
		ui.App.QueueUpdateDraw(func() {
			ui.StatusWindow.SetText(GetUpdateString())
		})
	}
}

var sp = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
var spCounter = 0

func GetUpdateString() string {
	s := fmt.Sprintf("%s Running | Signatures: %d | Temp work dir: %s", sp[7-spCounter%8], len(session.Signatures), *session.Options.TempDirectory)
	spCounter += 1
	return s
}
