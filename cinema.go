package claquete

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/pkg/errors"
)

type (
	// Cinema TODO
	Cinema struct {
		c           *Claquete
		ID          int    `json:"id"`
		Name        string `json:"name"`
		AddressLine string `json:"address_line"`
		TimeZone    string `json:"time_zone"`
	}
)

// GetCinema TODO
func GetCinema(id int) (*Cinema, error) {
	if id < 0 {
		return nil, errors.New("invalid ID")
	}

	idStr := strconv.Itoa(id)
	if idStr == "" {
		return nil, fmt.Errorf("couldn't create ID from integer %d", id)
	}

	// Uses cinema id in container class
	container := fmt.Sprintf("body > div.conteudo > div.progrb > div.cinema%d", id)

	var result *Cinema
	var err1 error

	c := NewClaquete()
	c.collector.OnHTML(container, func(e *colly.HTMLElement) {
		cinema, err := parseCinema(e.DOM)
		if err != nil {
			err1 = err
			return
		}
		cinema.c = c
		cinema.ID = id
		result = cinema
	})
	// Try to retrieve time zone
	c.collector.OnHTML("body > div.conteudo > div.progrb > div:nth-child(1) > p", func(e *colly.HTMLElement) {
		text := strings.TrimSpace(strings.Replace(e.Text, "por cinemas em", "", -1))
		if text != "" { // Expected state name
			result.TimeZone = getTimeZone(text)
		}
	})

	c.collector.OnScraped(func(*colly.Response) {
		if result == nil {
			err1 = errors.New("cinema not found")
		}
	})

	// The request is successful with any string
	// followed by .html
	u := fmt.Sprintf("%s/programacao/%s/.html", BaseURL, idStr)

	err := c.collector.Visit(u)
	if err == nil && err1 != nil {
		err = err1
	}

	return result, err
}

// GetNowPlaying retrieves now playing movies for Cinema.
// To get additional movie metadata use the GetMovie(int)
// function passing the retrieved movie id.
func (c *Cinema) GetNowPlaying() ([]Movie, error) {
	params := map[string]string{"cinema": strconv.Itoa(c.ID)}
	return getNowPlayingList(c.c.collector, "escolherFilme.php", params)
}

// GetCinemas TODO
func (c *Claquete) GetCinemas() ([]Cinema, error) {
	return getCinemas(c)
}

func getCinemas(c *Claquete) ([]Cinema, error) {
	var result []Cinema
	var err error

	if c.city == "" {
		return nil, errors.New("city was not specified")
	}

	c.collector.OnHTML("option", func(e *colly.HTMLElement) {
		value := e.Attr("value")

		ID, err := strconv.Atoi(value)
		if err != nil {
			err = errors.Wrapf(err, "conversion value %s to integer failed", value)
			return
		}
		if ID != 0 {
			result = append(result, Cinema{
				c:    c,
				ID:   ID,
				Name: e.Text,
			})
		}
	})

	u := fmt.Sprintf("%sescolherCinema_load.php", BaseAJAX)
	err = c.collector.Post(u, map[string]string{"cidade": c.city})
	return result, err
}

// GetCinemas retrieve cinema list
func GetCinemas(fu, city string) ([]Cinema, error) {
	return getCinemas(NewClaquete(
		FederativeUnit(fu),
		CityName(city),
	))
}

// GetSchedule ...
func (c *Cinema) GetSchedule() (*Schedule, error) {
	return GetSchedule(c.ID)
}

func parseCinema(s *goquery.Selection) (*Cinema, error) {
	addressLine := strings.TrimSpace(s.Find("div > div > span").Text())
	if addressLine != "" {
		addressLine = strings.Replace(addressLine, "\n", "", -1)
		addressLine = strings.Replace(addressLine, "(mapa)", "", -1)
	}
	result := &Cinema{
		Name:        strings.TrimSpace(s.Find("div > div > div.ttcine > h2").Text()),
		AddressLine: strings.TrimSpace(addressLine),
	}
	return result, nil
}

func setTime(t *time.Time, hours, minutes int) *time.Time {
	result := *t
	result = result.
		Add(time.Duration(hours) * time.Hour).
		Add(time.Duration(minutes) * time.Minute)
	return &result
}
