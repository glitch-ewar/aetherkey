package core

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
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

var tui UI
var signatures map[string][]MatchEvent

func GetUI() *UI {
	return &tui
}

func (ui *UI) Initialize() {
	signatures = make(map[string][]MatchEvent)

	ui.App = tview.NewApplication()

	ui.StatusWindow = tview.NewTextView()
	ui.StatusWindow.SetBorder(true)
	ui.StatusWindow.SetTitle("[#00FFFF::b]AETHERKEY")
	ui.StatusWindow.SetBorderColor(tcell.ColorAqua)
	ui.StatusWindow.SetDynamicColors(true)
	ui.StatusWindow.SetChangedFunc(func() {
		ui.App.Draw()
	})

	ui.LogWindow = tview.NewTextView()
	ui.LogWindow.SetBorder(true)
	ui.LogWindow.SetTitle("[::b]Log")
	ui.LogWindow.SetChangedFunc(func() {
		ui.App.Draw()
	})

	ui.SignaturesWindow = tview.NewList()
	ui.SignaturesWindow.SetBorder(true)
	ui.SignaturesWindow.SetTitle("[::b]Signatures")
	ui.SignaturesWindow.ShowSecondaryText(false)
	ui.SignaturesWindow.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		ui.DetailsWindow.Clear()

		signature := mainText
		columns := session.GetView(signature)
		for i, c := range columns {
			ui.DetailsWindow.SetCell(0, i, tview.NewTableCell(fmt.Sprintf("[::b]%s", c)))
		}

		for _, r := range signatures[signature] {
			ui.AddToDetailsWindow(signature, &r)
		}
	})

	ui.DetailsWindow = tview.NewTable()
	ui.DetailsWindow.SetBorders(true).
		SetBorder(true).
		SetTitle("[::b]Details")

	hflex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(ui.SignaturesWindow, 0, 1, false).
		AddItem(ui.DetailsWindow, 0, 6, false)

	ui.MainWindow = tview.NewFlex().SetDirection(tview.FlexRow)
	ui.MainWindow.AddItem(ui.StatusWindow, 3, 1, false)
	ui.MainWindow.AddItem(hflex, 0, 1, false)
	ui.MainWindow.AddItem(ui.LogWindow, 10, 1, false)

	go ui.UpdateStatus()
}

func (ui *UI) AddToDetailsWindow(signature string, event *MatchEvent) {
	selectedSignature, _ := ui.SignaturesWindow.GetItemText(ui.SignaturesWindow.GetCurrentItem())
	if selectedSignature == signature {
		event.AdditionalInfo["URL"] = event.Url
		idx := ui.DetailsWindow.GetRowCount()
		columns := session.GetView(signature)

		for i, column := range columns {
			value, exists := event.AdditionalInfo[column]
			textColor := ui.relevanceToColor(event.Relevance)
			if exists {
				ui.DetailsWindow.SetCell(idx, i, tview.NewTableCell(value).SetTextColor(textColor))
			} else {
				ui.DetailsWindow.SetCell(idx, i, tview.NewTableCell("???").SetTextColor(textColor))
			}
		}
	}
}

func (ui *UI) relevanceToColor(relevance Relevance) tcell.Color {
	if relevance == RelevanceHigh {
		return tcell.ColorAqua
	} else if relevance == RelevanceLow {
		return tcell.ColorGray
	} else {
		return tcell.ColorWhite
	}
}

func (ui *UI) Publish(event *MatchEvent) {
	results, contains := signatures[event.Signature]

	if !contains {
		results = []MatchEvent{}
		ui.SignaturesWindow.AddItem(event.Signature, "", 0, nil)
	}

	duplicate := false
	for _, r := range results {
		if r.Url == event.Url && r.Match == event.Match {
			duplicate = true
			break
		}
	}

	if !duplicate {
		signatures[event.Signature] = append(results, *event)
		ui.AddToDetailsWindow(event.Signature, event)
	}
}

func (ui *UI) Run() {
	if err := ui.App.SetRoot(ui.MainWindow, true).SetFocus(ui.SignaturesWindow).Run(); err != nil {
		panic(err)
	}
}

var spinner = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
var spinnerCounter = 0

func (ui *UI) UpdateStatus() {
	const refreshInterval = 500 * time.Millisecond
	for {
		time.Sleep(refreshInterval)
		ui.App.QueueUpdateDraw(func() {
			ui.StatusWindow.SetText(GetUpdateString())
			spinnerCounter += 1
		})
	}
}

func getSpinnerCharacter() string {
	return spinner[7-spinnerCounter%8]
}

func GetUpdateString() string {
	hideLowRelevance := false
	hideLowRelevanceText := "✕"
	if hideLowRelevance {
		hideLowRelevanceText = "✓"
	}

	return fmt.Sprintf("[#00FFFF]%s [::l]Running[::-] %s Signatures: %d | [::bu]H[::-]ide low relevance: %s | [::bu]Q[::-]uit",
		getSpinnerCharacter(),
		getSpinnerCharacter(),
		len(session.Signatures),
		hideLowRelevanceText)
}
