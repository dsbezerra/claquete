package claquete

import (
	"fmt"
	"log"
	"strings"

	"github.com/dsbezerra/claqueteapi/movieutil"
	"github.com/dsbezerra/claqueteapi/util"
	"github.com/gocolly/colly"
	"github.com/pkg/errors"
)

const (
	// BaseURL is Claquete's base website url
	BaseURL = "http://claquete.com.br"
	// BaseAJAX is the base path used to make AJAX requests
	BaseAJAX = BaseURL + "/lib/ajax/ajax."
)

type (
	// Claquete struct
	Claquete struct {
		FederativeUnit string
		CityName       string

		collector *colly.Collector
	}
)

// NewClaquete creates a new Claquete instance
func NewClaquete(options ...func(*Claquete)) *Claquete {
	c := &Claquete{}
	c.Init()

	for _, f := range options {
		f(c)
	}

	if c.FederativeUnit != "" {
		// Make sure we got the cookies
		c.GetCities()
	}

	return c
}

// Init initializes claquete's struct
func (c *Claquete) Init() {
	c.collector = colly.NewCollector(
		colly.UserAgent(util.RandomUserAgent()),
	)
}

// FederativeUnit sets the federative unit used by the Claquete.
func FederativeUnit(fu string) func(*Claquete) {
	if !isFederativeUnitValid(fu) {
		log.Fatal(fmt.Errorf("federative unit %s is invalid", fu))
	}
	return func(c *Claquete) {
		c.FederativeUnit = fu
	}
}

// CityName sets the city name used by the Claquete.
func CityName(city string) func(*Claquete) {
	if city == "" {
		log.Fatal(errors.New("city name is invalid"))
	}
	return func(c *Claquete) {
		c.CityName = city
	}
}

// GetReleases get releases of week
func (c *Claquete) GetReleases() ([]Movie, error) {
	var result []Movie
	var err error

	c.collector.OnHTML("#carrossel > ul:nth-child(1) > li", func(e *colly.HTMLElement) {
		m := Movie{
			Page:   e.DOM.Find("a").AttrOr("href", ""),
			Title:  strings.TrimSpace(e.DOM.Find("p").Text()),
			Poster: e.DOM.Find("img").AttrOr("src", ""),
		}

		ID, err := movieutil.IDFromURLString(m.Page)
		if err != nil {
			err = errors.Wrapf(err, "get url from %s failed", m.Page)
			return
		}

		m.ID = ID

		// TODO: check for missing fields
		result = append(result, m)
	})

	err = c.collector.Visit(BaseURL + "/noticias.html")

	return result, err
}
