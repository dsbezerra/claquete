package claquete

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dsbezerra/claqueteapi/util"
	"github.com/gocolly/colly"
)

// SearchFilterFlag is used to filter search results
type SearchFilterFlag uint8

const (
	// SearchFilterCinema restrict to Cinema results only
	SearchFilterCinema SearchFilterFlag = 1 << iota

	// SearchFilterMovie restrict to Movie results only
	SearchFilterMovie

	// SearchFilterNews restrict to News results only
	SearchFilterNews
)

const (
	// SearchTypeCinema indicates search result type is Cinema
	SearchTypeCinema = "cinema"

	// SearchTypeMovie indicates search result type is Movie
	SearchTypeMovie = "movie"

	// SearchTypeNews indicates search result type is News
	SearchTypeNews = "news"

	// DefaultSearchFilterFlags is the default behavior for search requests in Claquete's website.
	DefaultSearchFilterFlags = SearchFilterCinema | SearchFilterMovie | SearchFilterNews

	// MinQueryLength is used to avoid searching with a short query.
	// The website doesn't seem to have a pagination system and that makes
	// it retrieves all data in a single request.
	MinQueryLength = 2
)

var (
	// Months is an array of month names (pt-BR)
	Months = []string{
		"janeiro",
		"fevereiro",
		"março",
		"abril",
		"maio",
		"junho",
		"julho",
		"agosto",
		"setembro",
		"outubro",
		"novembro",
		"dezembro",
	}
)

type (
	// SearchResults represents the search operation results.
	SearchResults struct {
		FilterFlags SearchFilterFlag `json:"-"`

		Query      string         `json:"query"`
		Filtered   bool           `json:"filtered"`
		TotalCount int            `json:"total_count"`
		Results    []SearchResult `json:"results"`
	}

	// SearchResult represents one row returned from search operation.
	SearchResult struct {
		Date  time.Time `json:"date,omitempty"`
		Year  int       `json:"year,omitempty"`
		Title string    `json:"title"`
		Type  string    `json:"type"`
		Page  string    `json:"page"`
	}
)

// ShouldIncludeType checks if the result type match the filter.
func (s *SearchResults) ShouldIncludeType(t string) bool {
	var f SearchFilterFlag

	switch t {
	case SearchTypeCinema:
		f = SearchFilterCinema
	case SearchTypeMovie:
		f = SearchFilterMovie
	case SearchTypeNews:
		f = SearchFilterNews
	}

	return s.FilterFlags&f != 0
}

// Search performs a search operation without filtering results.
func Search(query string) (*SearchResults, error) {
	return search(query, DefaultSearchFilterFlags)
}

// SearchCinemas performs a search operation filtering results to Cinema only.
func SearchCinemas(query string) (*SearchResults, error) {
	return search(query, SearchFilterCinema)
}

// SearchMovies performs a search operation filtering results to Movie only.
func SearchMovies(query string) (*SearchResults, error) {
	return search(query, SearchFilterMovie)
}

// SearchNews performs a search operation filtering results to News only.
func SearchNews(query string) (*SearchResults, error) {
	return search(query, SearchFilterNews)
}

// SearchFiltered performs a search operation and applies a filter to its results.
// filterFlags is used to specify what it should return.
func SearchFiltered(query string, filterFlags SearchFilterFlag) (*SearchResults, error) {
	return search(query, filterFlags)
}

// Search searches in Claquete's website.
func search(query string, filterFlags SearchFilterFlag) (*SearchResults, error) {
	if len(query) < MinQueryLength {
		return nil, errors.New("query length must be at least 2")
	}

	result := &SearchResults{
		FilterFlags: filterFlags,
		Query:       query,
	}

	c := NewClaquete()
	c.collector.OnHTML("#busca_ajax", func(e *colly.HTMLElement) {
		totalCountRE := regexp.MustCompile("(\\d+)\\sresultados")
		var sr SearchResult
		e.DOM.Children().Each(func(i int, s *goquery.Selection) {
			// Parse found results count
			if s.Is("h3") && result.TotalCount == 0 {
				r := totalCountRE.FindStringSubmatch(strings.ToLower(s.Text()))
				if len(r) == 2 {
					totalCount, err := strconv.Atoi(r[1])
					if err != nil {
						fmt.Printf("couldn't find total count in text: %s\n", r[1])
					} else {
						result.TotalCount = totalCount
					}
				} else {
					fmt.Printf("expected size 2, got: %d\n", len(r))
				}
			} else if s.Is("div") && s.AttrOr("class", "") == "ttsubn" {
				t := getSearchType(s.Text())
				if result.ShouldIncludeType(t) {
					sr = SearchResult{
						Type: getSearchType(s.Text()),
					}
				}
			} else if s.Is("span") && sr.Type != "" {
				// If is a span and we have a result, parse the date/year information
				str := strings.TrimSpace(s.Text())
				if sr.Type == SearchTypeNews {
					d, _, err := util.CreateDate(str, " de ")
					if err != nil {
						// TODO: proper error handling
						fmt.Println(err)
					} else {
						sr.Date = d
					}
				} else if sr.Type == SearchTypeMovie && sr.Type != "" {
					year, err := strconv.Atoi(str)
					if err != nil {
						// TODO: proper error handling
						fmt.Printf("couldn't convert text %s to year\n", str)
					} else {
						sr.Year = year
					}
				} else if sr.Type == SearchTypeCinema && sr.Type != "" {
					// Do nothing.
				}
			} else if s.Is("h2") && sr.Type != "" {
				// If is a h2 and we have a result, get title and page information
				sr.Title = s.Text()
				sr.Page = s.Find("a").AttrOr("href", "")
				result.Results = append(result.Results, sr)
			}
		})

		// Update totalCount to match filtered results
		if result.FilterFlags != DefaultSearchFilterFlags {
			result.TotalCount = len(result.Results)
			result.Filtered = true
		}
	})

	// NOTE(diego):
	// Claquete website sends x, y positions of mouse
	// at the time search button was clicked.
	//
	// Request works without them, but let's send it anyway.
	searchBtnWidth := 30
	searchBtnHeight := 31
	err := c.collector.Post(BaseURL+"/busca.html", map[string]string{
		"query": query,
		"x":     strconv.Itoa(rand.Intn(searchBtnWidth)),
		"y":     strconv.Itoa(rand.Intn(searchBtnHeight)),
	})
	return result, err
}

// getSearchType retrieves the search type for a given string.
func getSearchType(str string) string {
	result := ""
	str = strings.TrimSpace(str)
	switch strings.ToLower(str) {
	case "cinema":
		result = SearchTypeCinema
	case "filme":
		result = SearchTypeMovie
	case "notícias":
		result = SearchTypeNews
	}
	return result
}
