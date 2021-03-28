package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/stephenwithav/go-hmmm"
)

// filterOutSeenPapers ...
func filterOutSeenPapers(seenURLs map[string]interface{}, potentialPapers []hmmm.Paper) []hmmm.Paper {
	var uniquePapers []hmmm.Paper
	for _, paper := range potentialPapers {
		if _, ok := seenURLs[paper.ArticleID]; ok {
			continue
		}

		uniquePapers = append(uniquePapers, paper)
		seenURLs[paper.ArticleID] = true
	}

	return uniquePapers
}

// retrieveSectionsFromWeb ...
func retrieveSectionsFromWeb(secs []string) ([]hmmm.Paper, error) {
	var uniquePapers []hmmm.Paper
	seenAlready := map[string]interface{}{}
	for _, s := range secs {
		n, err := hmmm.CountNewPapersFromArxiv(s)
		if err != nil {
			return nil, err
		}
		papers, err := hmmm.GetPapersFromArxiv(n, s)
		if err != nil {
			return nil, err
		}

		newPapers := filterOutSeenPapers(seenAlready, papers)
		uniquePapers = append(uniquePapers, hmmm.Paper{Title: fmt.Sprintf("-- %s [%d; %d are already listed above] ------------", s, len(newPapers), len(papers)-len(newPapers))})
		uniquePapers = append(uniquePapers, newPapers...)
	}

	return uniquePapers, nil
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

	sections := []string{"q-fin", "cs.AI", "cs.LG", "cs.CV"}
	if len(os.Args) > 1 {
		sections = os.Args[1:]
	}
	papers, err := retrieveSectionsFromWeb(sections)
	if err != nil {
		fmt.Printf("Error retrieving the latest articles from [%s]: %v\n", sections, err)
		os.Exit(3)
	}

	papersListFormat := "Papers from %s - [%d/%d]"
	starsListFormat := "Starred Papers from %s - [%d/%d]"

	inactiveStyle := ui.NewStyle(ui.ColorWhite)
	activeStyle := ui.NewStyle(ui.ColorYellow)

	papersList := widgets.NewList()
	papersList.WrapText = false
	papersList.TextStyle = inactiveStyle
	papersList.SelectedRowStyle = activeStyle
	papersList.Rows = paperTitles(papers)
	papersList.Title = fmt.Sprintf(papersListFormat, sections, papersList.SelectedRow+1, len(papersList.Rows))

	starsList := widgets.NewList()
	starsList.WrapText = false
	starsList.TextStyle = inactiveStyle
	starsList.Rows = []string{}
	starsList.Title = fmt.Sprintf(starsListFormat, sections, 0, 0)

	abstractView := widgets.NewParagraph()
	abstractView.WrapText = true
	abstractView.TextStyle = inactiveStyle

	rowsToExtract := []int{}

	starsGrid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	starsGrid.SetRect(0, 0, termWidth, termHeight)

	starsGrid.Set(
		ui.NewRow(0.75, papersList),
		ui.NewRow(0.25, starsList),
	)

	activeGrid := starsGrid

	ui.Render(activeGrid)

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
				ui.Render(activeGrid)
				time.Sleep(10 * time.Second)

				body := cfg.GetString("Body")
				body = strings.ReplaceAll(body, "%TITLE%", papers[row].Title)
				body = strings.ReplaceAll(body, "%URL%", papers[row].ScienceWiseURL())
				initialTweet, _, err = sendTweet(client, body, initialTweet.ID)
				if err != nil {
					fmt.Printf("Unable to add [%s] to thread. : %#v\n", papers[row].Title, err)
				}

				if (i+2)%200 == 0 {
					time.Sleep(3 * time.Hour)
				}
			}

			_, _, err = sendTweet(client, "@threadreaderapp unroll", initialTweet.ID)
		case "p":
			if activeList == papersList {
				list := starsList.Rows

				abstract, err := hmmm.GetAbstractFromPaper(papers[papersList.SelectedRow])
				if err != nil {
					log.Fatal(err)
				}
				starsList.Rows = []string{abstract}
				starsList.WrapText = true
				ui.Render(activeGrid)

				e = <-uiEvents

				starsList.Rows = list
				starsList.WrapText = false
			}
		}

		if previousKey == "g" {
			previousKey = ""
		} else {
			previousKey = e.ID
		}

		papersList.Title = fmt.Sprintf(papersListFormat, sections, papersList.SelectedRow+1, len(papersList.Rows))
		if len(starsList.Rows) > 0 {
			starsList.Title = fmt.Sprintf(starsListFormat, sections, starsList.SelectedRow+1, len(starsList.Rows))
		}
		ui.Render(activeGrid)
	}
}
