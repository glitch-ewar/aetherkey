package core

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

type UI struct {
	App          *tview.Application
	MainWindow   *tview.Flex
	StatusWindow *tview.TextView
	LogWindow    *tview.TextView
}

var tui UI

func GetUI() *UI {
	return &tui
}

func (ui *UI) Initialize() {
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

	hflex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Left"), 0, 1, false).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Middle (3 x height of Top)"), 0, 6, false)

	ui.MainWindow = tview.NewFlex().SetDirection(tview.FlexRow)
	ui.MainWindow.AddItem(ui.StatusWindow, 3, 1, false)
	ui.MainWindow.AddItem(hflex, 0, 1, false)
	ui.MainWindow.AddItem(ui.LogWindow, 8, 1, false)

	go ui.UpdateStatus()
}

func (ui *UI) Run() {
	if err := ui.App.SetRoot(ui.MainWindow, true).SetFocus(ui.MainWindow).Run(); err != nil {
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
