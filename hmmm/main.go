package main

import (
	"fmt"
	"os"
	"strings"
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
	cfg, err := getConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := ui.Init(); err != nil {
		fmt.Printf("failed to initialize termui: %v\n", err)
		os.Exit(2)
	}
	defer ui.Close()

	section := "cs.LG"
	if len(os.Args) > 1 {
		section = os.Args[1]
	}
	papers, err := retrieveSectionFromWeb(section)
	if err != nil {
		fmt.Printf("Error retrieving the latest articles from [%s]: %v\n", section, err)
		os.Exit(3)
	}

	papersListFormat := "Papers from [%s] - [%d/%d]"
	starsListFormat := "Starred Papers from [%s] - [%d/%d]"

	inactiveStyle := ui.NewStyle(ui.ColorWhite)
	activeStyle := ui.NewStyle(ui.ColorClear)

	papersList := widgets.NewList()
	papersList.WrapText = false
	papersList.TextStyle = inactiveStyle
	papersList.SelectedRowStyle = activeStyle
	papersList.Rows = paperTitles(papers)
	papersList.Title = fmt.Sprintf(papersListFormat, section, papersList.SelectedRow+1, len(papersList.Rows))

	starsList := widgets.NewList()
	starsList.WrapText = false
	starsList.TextStyle = inactiveStyle
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
				papersList.SelectedRowStyle = inactiveStyle
				starsList.SelectedRowStyle = activeStyle
				activeList = starsList
			} else {
				starsList.SelectedRowStyle = inactiveStyle
				papersList.SelectedRowStyle = activeStyle
				activeList = papersList
			}
		case "<C-t>":
			client := createTwitterOathClient(cfg.GetString("ConsumerKey"), cfg.GetString("ConsumerSecret"), cfg.GetString("AccessToken"), cfg.GetString("AccessSecret"))
			initialTweet, _, err := sendTweet(client, cfg.GetString("Intro"), 0)
			if err != nil {
				fmt.Printf("Unable to initialize thread: %#v\n", err)
				os.Exit(4)
			}

			papersList.SelectedRowStyle = inactiveStyle
			starsList.SelectedRowStyle = activeStyle
			for i, row := range rowsToExtract {
				starsList.SelectedRow = i
				starsList.Title = fmt.Sprintf("Tweeting paper %d of %d", i+1, len(rowsToExtract))
				ui.Render(grid)
				time.Sleep(10 * time.Second)

				body := cfg.GetString("Body")
				body = strings.ReplaceAll(body, "%TITLE%", papers[row].Title)
				body = strings.ReplaceAll(body, "%URL%", papers[row].ScienceWiseURL())
				initialTweet, _, err = sendTweet(client, body, initialTweet.ID)
				if err != nil {
					fmt.Printf("Unable to add [%s] to thread. : %#v\n", papers[row].Title, err)
				}
			}
		}

		if previousKey == "g" {
			previousKey = ""
		} else {
			previousKey = e.ID
		}

		papersList.Title = fmt.Sprintf(papersListFormat, section, papersList.SelectedRow+1, len(papersList.Rows))
		if len(starsList.Rows) > 0 {
			starsList.Title = fmt.Sprintf(starsListFormat, section, starsList.SelectedRow+1, len(starsList.Rows))
		}
		ui.Render(grid)
	}
}
