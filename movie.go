package claquete

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/dsbezerra/claqueteapi/movieutil"
	"github.com/dsbezerra/claqueteapi/util"
	"github.com/gocolly/colly"
	"github.com/pkg/errors"
)

const (
	// RatingL is the url path to L or Livre rating image
	RatingL = "img/icos/l.png"
	// Rating10 is the url path to age 10 rating image
	Rating10 = "img/icos/10.png"
	// Rating12 is the url path to age 12 rating image
	Rating12 = "img/icos/12.png"
	// Rating14 is the url path to age 14 rating image
	Rating14 = "img/icos/14.png"
	// Rating16 is the url path to age 16 rating image
	Rating16 = "img/icos/16.png"
	// Rating18 is the url path to age 18 rating image
	Rating18 = "img/icos/18.png"

	// ImageTypeAny used to classify images in gallery
	ImageTypeAny = "any"
	// ImageTypePoster used to classify poster images in gallery
	ImageTypePoster = "poster"
)

var (
	errMovieNotFound = errors.New("movie not found")
)

type (

	// Movie represents a movie in Claquete's website
	Movie struct {
		ID            int        `json:"id"`
		Page          string     `json:"page"`
		Slug          string     `json:"slug,omitempty"`
		Title         string     `json:"title,omitempty"`
		OriginalTitle string     `json:"original_title,omitempty"`
		Synopsis      string     `json:"synopsis,omitempty"`
		Poster        string     `json:"poster,omitempty"`
		Country       string     `json:"country,omitempty"`
		Distributor   string     `json:"distributor,omitempty"`
		Genres        []string   `json:"genres,omitempty"`
		Cast          []string   `json:"cast,omitempty"`
		Screenplay    []string   `json:"screenplay,omitempty"`
		Production    []string   `json:"production,omitempty"`
		Direction     []string   `json:"direction,omitempty"`
		Runtime       int        `json:"runtime,omitempty"`
		Rating        int        `json:"rating,omitempty"`
		ReleaseDate   *time.Time `json:"release_date,omitempty"`
		Images        []Image    `json:"images,omitempty"`
	}

	// Image TODO
	Image struct {
		URL        string `json:"url"`
		Type       string `json:"type"`
		Width      int    `json:"width"`
		Height     int    `json:"height"`
		Resolution int    `json:"resolution"`
		Format     string `json:"format"`
	}
)

// GetMovie retrieves movie information for the given ID
func GetMovie(id int) (*Movie, error) {
	if id < 0 {
		return nil, errors.New("invalid ID")
	}

	idStr := strconv.Itoa(id)
	if idStr == "" {
		return nil, fmt.Errorf("couldn't create ID from integer %d", id)
	}

	var result *Movie
	var err error

	c := NewClaquete()
	c.collector.OnResponse(func(r *colly.Response) {
		// Make sure we got a movie page
		ID, err := movieutil.IDFromURL(r.Request.URL)
		if err != nil {
			fmt.Println(err)
			c.collector.OnHTMLDetach("div.mvposter")
			c.collector.OnHTMLDetach("div.mvposter > div.mvclassif img")
			c.collector.OnHTMLDetach("div.mvdesc")
			c.collector.OnHTMLDetach("#cont1")
			c.collector.OnHTMLDetach("#cont2 > galeria img")
		} else {
			result = &Movie{
				ID:   ID,
				Page: r.Request.URL.String(),
			}
		}
	})

	c.collector.OnHTML("div.mvposter", func(e *colly.HTMLElement) {
		result.Poster = e.Attr("src")
	})

	c.collector.OnHTML("div.mvposter > div.mvclassif img", func(e *colly.HTMLElement) {
		src := e.Attr("src")
		if src != "" {
			if strings.Contains(src, RatingL) {
				result.Rating = -1
			} else if strings.Contains(src, Rating10) {
				result.Rating = 10
			} else if strings.Contains(src, Rating12) {
				result.Rating = 12
			} else if strings.Contains(src, Rating14) {
				result.Rating = 14
			} else if strings.Contains(src, Rating16) {
				result.Rating = 16
			} else if strings.Contains(src, Rating18) {
				result.Rating = 18
			}
		}
	})

	c.collector.OnHTML("div.mvdesc", func(e *colly.HTMLElement) {
		result.Title = strings.TrimSpace(e.DOM.Find("h1").Text())
		result.Slug = util.CreateSlug(result.Title)
		ot := strings.TrimSpace(e.DOM.Find("h2").Text())
		if ot != "" {
			// Skip parentheses and year
			// Ex: (Original Title, Year)
			start := 1
			end := strings.LastIndex(ot, ",")
			if end > -1 {
				value := ot[start:end]
				if value != "Ainda Sem Título em Português" {
					result.OriginalTitle = value
				}
			}
		}

		e.DOM.Find("p").Each(func(i int, s *goquery.Selection) {
			label, value := util.BreakByToken(s.Text(), ':')
			label = strings.ToLower(label)
			if label != "" && value != "" {
				switch label {
				case "país":
					result.Country = value

				case "gênero":
					result.Genres = strings.Split(value, ", ")

				case "duração":
					runtime, err := movieutil.ParseRuntime(value)
					if err != nil {
						fmt.Println(err)
					} else {
						result.Runtime = runtime
					}

				case "distr.":
					result.Distributor = value

				case "estreia", "estreia.":
					// NOTE(diego):  This ^ dot exists in website and is probably a
					// typo or cut and paste from distr.
					rd, err := movieutil.ParseReleaseDate(value, "/")
					if err != nil {
						fmt.Println(err)
					} else {
						result.ReleaseDate = rd
					}
				}
			}
		})
	})

	c.collector.OnHTML("#cont1", func(e *colly.HTMLElement) {
		getSplitted := func(str string) []string {
			str = strings.TrimFunc(str, func(r rune) bool {
				return unicode.IsSpace(r) || r == ','
			})
			str = strings.Replace(str, "\n", " ", -1)
			return strings.Split(str, ", ")
		}
		e.DOM.Children().Each(func(i int, s *goquery.Selection) {
			if s.Is("h3") {
				header := strings.TrimSpace(strings.ToLower(s.Text()))
				value := strings.TrimSpace(s.Next().Text())
				switch header {
				case "sinopse":
					result.Synopsis = value
				case "elenco":
					result.Cast = getSplitted(value)
				case "roteiro":
					result.Screenplay = getSplitted(value)
				case "produção":
					result.Production = getSplitted(value)
				case "direção":
					result.Direction = getSplitted(value)
				}
			}
		})
	})

	c.collector.OnHTML("#cont2", func(e *colly.HTMLElement) {
		e.DOM.Find("img").Each(func(i int, s *goquery.Selection) {
			src := s.AttrOr("src", "")
			if src != "" {
				image, err := util.GetImage(src)
				if err != nil {
					fmt.Printf("error %s ocurred while getting image from %s\n", err.Error(), src)
				} else {
					i := Image{
						URL:        src,
						Width:      image.Width,
						Height:     image.Height,
						Resolution: image.Resolution,
						Format:     image.Format,
					}
					// Likely to be a poster image.
					if image.Height > image.Width {
						i.Type = "poster"
					} else {
						i.Type = ImageTypeAny
					}
					result.Images = append(result.Images, i)
				}
			}
		})
	})

	err = c.collector.Visit(BaseURL + "/filmes/filme.php?cf=" + idStr)

	// Ignore movies without title.
	if isMovieInvalid(result) {
		err = errMovieNotFound
		result = nil
	}

	return result, err
}

// getNowPlayingList is a helper to retrieve list of
// movies inside a select element
func getNowPlayingList(c *colly.Collector, path string, params map[string]string) ([]Movie, error) {
	var result []Movie
	var err error
	c.OnHTML("option", func(e *colly.HTMLElement) {
		value := e.Attr("value")
		ID, err := strconv.Atoi(value)
		if err != nil {
			fmt.Println(errors.Wrapf(err, "conversion value %s to integer failed", value))
			return
		}
		if ID != 0 {
			result = append(result, Movie{
				ID:    ID,
				Title: e.Text,
				Page:  fmt.Sprintf("%s/filmes/filme.php?cf=%d", BaseURL, ID),
			})
		}
	})
	err = c.Post(BaseAJAX+path, params)
	return result, err
}

func isMovieInvalid(m *Movie) bool {
	if m == nil {
		return false
	}
	return m.Title == "" &&
		m.OriginalTitle == "" &&
		m.Synopsis == "" &&
		m.Poster == "" &&
		m.Rating == 0 &&
		m.ReleaseDate == nil &&
		m.Slug == "" &&
		m.Country == "" &&
		m.Distributor == "" &&
		len(m.Genres) == 0 &&
		len(m.Cast) == 0 &&
		len(m.Production) == 0 &&
		len(m.Direction) == 0 &&
		len(m.Screenplay) == 0
}
