package claquete

import (
	"fmt"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dsbezerra/claqueteapi/movieutil"
	"github.com/dsbezerra/claqueteapi/util"
	"github.com/gocolly/colly"
)

type (
	// ReleaseWeek struct
	ReleaseWeek struct {
		Date   time.Time `json:"date"`
		Movies []Movie   `json:"movies"`
	}

	// Calendar struct
	Calendar struct {
		Month time.Month    `json:"month"`
		Year  int           `json:"year"`
		Weeks []ReleaseWeek `json:"weeks"`
	}
)

// GetCalendar retrieves calendar for current month and year
func GetCalendar() (*Calendar, error) {
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	now := time.Now().In(loc)
	return getCalendar(now.Month(), now.Year())
}

// GetCalendarAt retrieves calendar for the given month and year
func GetCalendarAt(month time.Month, year int) (*Calendar, error) {
	return getCalendar(month, year)
}

func getCalendar(month time.Month, year int) (*Calendar, error) {
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	result := &Calendar{
		Month: month,
		Year:  year,
	}

	var week ReleaseWeek

	c := NewClaquete()
	c.collector.OnHTML("*", func(e *colly.HTMLElement) {
		e.DOM.Each(func(i int, s *goquery.Selection) {
			class := s.AttrOr("class", "")
			if class != "" {
				if s.Is("div") && class == "cxsem" {
					week = ReleaseWeek{}
					d := s.Find("p").Text()
					day, err := strconv.Atoi(d)
					if err != nil {
						fmt.Printf("couldn't convert day to integer in text %s", d)
					} else {
						week.Date = time.Date(result.Year, result.Month, day, 0, 0, 0, 0, loc)
					}
				} else if s.Is("ul") && class == "posters" {
					s.Children().Each(func(j int, ss *goquery.Selection) {
						movie := Movie{
							Title:  ss.Find("ul > li > p").Text(),
							Poster: ss.Find("img").AttrOr("src", ""),
							Page:   ss.Find("div[data-hint=\"Sinopse\"]").Parent().AttrOr("href", ""),
						}
						if movie.Page != "" {
							slug, err := movieutil.SlugFromURLString(movie.Page)
							if err != nil {
								fmt.Println(err)
								// Fallback to our slug creation function
								movie.Slug = util.CreateSlug(movie.Title)
							} else {
								movie.Slug = slug
							}
						}
						ss.Find("a").EachWithBreak(func(i int, s *goquery.Selection) bool {
							ID, err := movieutil.IDFromURLString(s.AttrOr("href", ""))
							if err == nil {
								movie.ID = ID
								return false
							}
							return true
						})
						week.Movies = append(week.Movies, movie)
					})
					result.Weeks = append(result.Weeks, week)
				}
			}
		})
	})

	err := c.collector.Post(BaseAJAX+"calendario.php", map[string]string{
		"ano": strconv.Itoa(year),
		"mes": strconv.Itoa(int(month)),
	})

	return result, err
}

// NextMonth get next month calendar from the current one
func (cal *Calendar) NextMonth() (*Calendar, error) {
	month, year := cal.Month, cal.Year
	month++
	if month > time.December {
		month = time.January
		year++
	}
	return getCalendar(month, year)
}

// PrevMonth get previous month calendar from the current one
func (cal *Calendar) PrevMonth() (*Calendar, error) {
	month, year := cal.Month, cal.Year
	month--
	if month < time.January {
		month = time.December
		year--
	}
	return getCalendar(month, year)
}
