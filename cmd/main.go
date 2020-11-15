package main

import (
	"fmt"
	"log"
	"os"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/stephenwithav/go-hmmm"
)

// retrieveSectionFromWeb ...
func retrieveSectionFromWeb(s string) ([]hmmm.Paper, error) {
	n, err := hmmm.CountNewPapersFromArxiv(s)
	if err != nil {
		return nil, err
	}
	papers, err := hmmm.GetPapersFromArxiv(n, s)

	return papers, err
}

// paperTitles ...
func paperTitles(papers []hmmm.Paper) []string {
	titles := []string{}
	for _, p := range papers {
		titles = append(titles, p.Title)
	}

	return titles
}

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	section := "cs.LG"
	if len(os.Args) > 1 {
		section = os.Args[1]
	}
	papers, err := retrieveSectionFromWeb(section)
	if err != nil {
		log.Fatal(err)
	}

	papersList := widgets.NewList()
	papersList.Title = fmt.Sprintf("Papers from [%s]", section)
	papersList.WrapText = false
	papersList.SetRect(0, 0, 40, 40)
	papersList.TextStyle = ui.NewStyle(ui.ColorWhite)
	papersList.SelectedRowStyle = ui.NewStyle(ui.ColorBlue)
	papersList.Rows = paperTitles(papers)

	starsList := widgets.NewList()
	starsList.Title = fmt.Sprintf("Starred [%s] papers", section)
	starsList.WrapText = false
	starsList.SetRect(0, 0, 40, 40)
	starsList.TextStyle = ui.NewStyle(ui.ColorWhite)
	starsList.Rows = []string{}

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		ui.NewRow(0.5, papersList),
		ui.NewRow(0.5, starsList),
	)

	ui.Render(grid)

	previousKey := ""
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		case "j", "<Down>":
			papersList.ScrollDown()
		case "k", "<Up>":
			papersList.ScrollUp()
		case "<C-d>":
			papersList.ScrollHalfPageDown()
		case "<C-u>":
			papersList.ScrollHalfPageUp()
		case "<C-f>":
			papersList.ScrollPageDown()
		case "<C-b>":
			papersList.ScrollPageUp()
		case "g":
			if previousKey == "g" {
				papersList.ScrollTop()
			}
		case "<Home>":
			papersList.ScrollTop()
		case "G", "<End>":
			papersList.ScrollBottom()
		case "s":
			starsList.Rows = append(starsList.Rows, papersList.Rows[papersList.SelectedRow])
		}

		if previousKey == "g" {
			previousKey = ""
		} else {
			previousKey = e.ID
		}

		ui.Render(grid)
	}
}
