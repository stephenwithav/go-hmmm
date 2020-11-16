package main

import (
	"fmt"
	"log"
	"os"
	"time"

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

	papersListFormat := "Papers from [%s] - [%d/%d]"
	starsListFormat := "Papers from [%s] - [%d/%d]"

	papersList := widgets.NewList()
	papersList.WrapText = false
	papersList.TextStyle = ui.NewStyle(ui.ColorWhite)
	papersList.SelectedRowStyle = ui.NewStyle(ui.ColorBlue)
	papersList.Rows = paperTitles(papers)
	papersList.Title = fmt.Sprintf(papersListFormat, section, papersList.SelectedRow+1, len(papersList.Rows))

	starsList := widgets.NewList()
	starsList.WrapText = false
	starsList.TextStyle = ui.NewStyle(ui.ColorWhite)
	starsList.Rows = []string{}
	starsList.Title = fmt.Sprintf(starsListFormat, section, 0, 0)

	rowsToExtract := []int{}

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		ui.NewRow(0.75, papersList),
		ui.NewRow(0.25, starsList),
	)

	ui.Render(grid)

	previousKey := ""
	uiEvents := ui.PollEvents()
	activeList := papersList
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		case "j", "<Down>":
			activeList.ScrollDown()
		case "k", "<Up>":
			activeList.ScrollUp()
		case "<C-d>":
			activeList.ScrollHalfPageDown()
		case "<C-u>":
			activeList.ScrollHalfPageUp()
		case "<C-f>":
			activeList.ScrollPageDown()
		case "<C-b>":
			activeList.ScrollPageUp()
		case "g":
			if previousKey == "g" {
				activeList.ScrollTop()
			}
		case "<Home>":
			activeList.ScrollTop()
		case "G", "<End>":
			activeList.ScrollBottom()
		case "s":
			if activeList == papersList {
				starsList.Rows = append(starsList.Rows, papersList.Rows[papersList.SelectedRow])
				rowsToExtract = append(rowsToExtract, papersList.SelectedRow)
				starsList.ScrollBottom()
			}
		case "u":
			if activeList == papersList {
				break
			}
			if starsList.SelectedRow == 0 {
				starsList.Rows = starsList.Rows[1:]
				rowsToExtract = rowsToExtract[1:]
			} else {
				rowsToExtract = append(rowsToExtract[0:starsList.SelectedRow], rowsToExtract[starsList.SelectedRow+1:]...)				
				starsList.Rows = append(starsList.Rows[0:starsList.SelectedRow], starsList.Rows[starsList.SelectedRow+1:]...)				
			}
		case "<C-i>", "<Tab>":
			if activeList == papersList {
				papersList.SelectedRowStyle = ui.NewStyle(ui.ColorWhite)
				starsList.SelectedRowStyle = ui.NewStyle(ui.ColorBlue)
				activeList = starsList
			} else {
				starsList.SelectedRowStyle = ui.NewStyle(ui.ColorWhite)
				papersList.SelectedRowStyle = ui.NewStyle(ui.ColorBlue)
				activeList = papersList
			}
		case "<C-e>": // Extract
			w, err:=os.Create(time.Now().Format("2006-01-02") + "-" + section + ".html")
			if err != nil {
				break
			}
			fmt.Fprint(w, "<html><body><textarea>\n")
			for _, row := range rowsToExtract {
				fmt.Fprintf(w, "%s\n\n%s\n\n", papers[row], papers[row].ScienceWiseURL())
			}
			fmt.Fprint(w, "\n</textarea></body></html>")
			w.Close()
		}

		if previousKey == "g" {
			previousKey = ""
		} else {
			previousKey = e.ID
		}

		papersList.Title = fmt.Sprintf(papersListFormat, section, papersList.SelectedRow+1, len(papersList.Rows))
		starsList.Title = fmt.Sprintf(starsListFormat, section, starsList.SelectedRow+1, len(starsList.Rows))
		ui.Render(grid)
	}
}
