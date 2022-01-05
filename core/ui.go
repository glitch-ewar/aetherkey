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

type Result map[string]string

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
	ui.SignaturesWindow.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		ui.DetailsWindow.Clear()

		signature := mainText
		processor := session.GetProcessor(signature)
		for i, c := range processor.Columns {
			ui.DetailsWindow.SetCell(0, i, tview.NewTableCell(c))
		}

		for _, r := range signatures[signature] {
			ui.AddToDetailsWindow(signature, r)
		}
	})

	ui.DetailsWindow = tview.NewTable()
	ui.DetailsWindow.SetBorders(true).
		SetBorder(true).
		SetTitle("Details")

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
	selectedSignature, _ := ui.SignaturesWindow.GetItemText(ui.SignaturesWindow.GetCurrentItem())
	if selectedSignature == signature {
		idx := ui.DetailsWindow.GetRowCount()
		processor := session.GetProcessor(signature)

		for i, c := range processor.Columns {
			ui.DetailsWindow.SetCell(idx, i, tview.NewTableCell(result[c]))
		}
	}
}

func (ui *UI) Publish(signature string, repository string, match string) {
	results, contains := signatures[signature]

	if !contains {
		results = make([]Result, 0)
		ui.SignaturesWindow.AddItem(signature, "", 0, nil)
	}

	result := make(map[string]string)
	result["Repository"] = repository
	result["Match"] = match

	duplicate := false
	for _, r := range results {
		if r["Repository"] == repository && r["Match"] == match {
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
