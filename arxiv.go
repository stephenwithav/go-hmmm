package hmmm

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const arxivTemplate = "https://arxiv.org/list/%s/pastweek"

type Paper struct {
	Title     string
	ArticleID string
}

func (p Paper) String() string {
	return fmt.Sprintf("[%s] %s", p.ArticleID, p.Title)
}

func (p Paper) ArxivURL() string {
	return fmt.Sprintf("http://arxiv.org/abs/%s", p.ArticleID)
}

func (p Paper) ScienceWiseURL() string {
	return fmt.Sprintf("http://sciencewise.info/bookmarks/%s/add", p.ArticleID)
}

func (p Paper) RedditURL() string {
	return fmt.Sprint("https://reddit.com/submit?url=" + url.QueryEscape(fmt.Sprintf("https://arxiv.org/abs/"+p.ArticleID)))
}

// CountNewPapersFromArxiv returns a string representation of an int,
// which specifies the number of new papers added in the past 7 days. 
//
// An error is returned for any errors that may occur along the way.
// (e.g., http, parsing)
func CountNewPapersFromArxiv(section string) (string, error) {
	res, err := http.Get(fmt.Sprintf(arxivTemplate, section))
	if err != nil {
		return "", fmt.Errorf("Unable to retrieve new paper count: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("Unable to retrieve new paper count: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", fmt.Errorf("Unable to extract new paper count: %s", err)
	}

	aLastChild := doc.Find("#dlpage > small > a:last-child").First()
	return strings.Split(aLastChild.Text(), "-")[1], nil // 1-N, return N.
}

// GetPapersFromReader returns a slice of new Paprs from the given
// io.Reader.
//
// An error is returned for any errors that may occur along the way.
// (e.g., http, parsing)
func GetPapersFromReader(r io.Reader) ([]Paper, error) {
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("Error detected while decoding io.Reader: %s", err)
	}

	var papers []Paper
	// Retrieve the full arXiv list.
	dl := doc.Find("#dlpage > dl")
	dl.Find("dt").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the arXiv id
		arxivTag := s.Find("a[title='Abstract']")
		id := strings.Split(arxivTag.Text(), ":")
		papers = append(papers, Paper{"", id[1]})
	})

	dl.Find("dd").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the arXiv id
		arxivTitle := s.Find("div.list-title").Text()
		arxivTitle = strings.Trim(arxivTitle, "\n")
		arxivTitle = strings.ReplaceAll(arxivTitle, "  ", " ")
		papers[i].Title = strings.TrimPrefix(arxivTitle, "Title: ")
	})

	return papers, nil
}

// GetPapersFromArxiv returns a slice of Papers, along with any error
// that may occur while retrieving the Paper metadata
func GetPapersFromArxiv(n, section string) ([]Paper, error) {
	// Request the HTML page.
	res, err := http.Get(fmt.Sprintf(arxivTemplate, section) + "?show=" + n)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	return GetPapersFromReader(res.Body)
}
