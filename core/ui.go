package core

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

type UI struct {
	App              *tview.Application
	MainWindow       *tview.Flex
	StatusWindow     *tview.TextView
	SignaturesWindow *tview.List
	DetailsWindow    *tview.Table
	LogWindow        *tview.TextView
}

type Result struct {
	Url   string
	Match string
}

var tui UI
var signatures map[string][]Result

func GetUI() *UI {
	return &tui
}

func (ui *UI) Initialize() {
	signatures = make(map[string][]Result)

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
	ui.SignaturesWindow.ShowSecondaryText(false)

	ui.DetailsWindow = tview.NewTable()
	ui.DetailsWindow.SetBorders(true).
		SetBorder(true).
		SetTitle("Details")

	ui.DetailsWindow.SetCell(0, 0, tview.NewTableCell("URL"))
	ui.DetailsWindow.SetCell(0, 1, tview.NewTableCell("Match"))

	hflex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(ui.SignaturesWindow, 0, 1, false).
		AddItem(ui.DetailsWindow, 0, 6, false)

	ui.MainWindow = tview.NewFlex().SetDirection(tview.FlexRow)
	ui.MainWindow.AddItem(ui.StatusWindow, 3, 1, false)
	ui.MainWindow.AddItem(hflex, 0, 1, false)
	ui.MainWindow.AddItem(ui.LogWindow, 10, 1, false)

	go ui.UpdateStatus()
}

func (ui *UI) AddToDetailsWindow(signature string, result Result) {
	idx := ui.DetailsWindow.GetRowCount()

	ui.DetailsWindow.SetCell(idx, 0, tview.NewTableCell(result.Url))
	ui.DetailsWindow.SetCell(idx, 1, tview.NewTableCell(result.Match))

}

func (ui *UI) Publish(signature string, url string, matches []string) {
	results, contains := signatures[signature]

	if !contains {
		results = make([]Result, 0)
		ui.SignaturesWindow.AddItem(signature, "", 0, nil)
	}

	for _, match := range matches {
		result := Result{Url: url, Match: match}
		duplicate := false
		for _, r := range results {
			if r.Url == url && r.Match == match {
				duplicate = true
				break
			}
		}

		if !duplicate {
			results = append(results, result)
			signatures[signature] = results
			ui.AddToDetailsWindow(signature, result)
		}
	}
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
