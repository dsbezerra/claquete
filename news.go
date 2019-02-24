package claquete

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dsbezerra/claqueteapi/util"
	"github.com/gocolly/colly"
	"github.com/pkg/errors"
)

type (
	// Headline struct
	Headline struct {
		Title    string     `json:"title"`
		Category string     `json:"category,omitempty"`
		Date     *time.Time `json:"date,omitempty,omitempty"`
		Image    string     `json:"image,omitempty"`
		NewsPage string     `json:"news_page,omitempty"`
	}
	// News struct
	News struct {
		Author   string   `json:"author"`
		Headline Headline `json:"headline"`
		Page     string   `json:"page"`
		Content  string   `json:"content"`
		HTML     string   `json:"html"`
	}
)

// GetHeadlines ...
func GetHeadlines() ([]Headline, error) {
	var result []Headline
	var err error

	c := NewClaquete()
	c.collector.OnHTML("body > div.conteudo > div.noticias", func(e *colly.HTMLElement) {
		h := Headline{}
		e.DOM.Children().Each(func(i int, s *goquery.Selection) {
			// Only in highlight
			if s.Is("a") {
				p := s.Find("a > div.principal")
				if p.Length() != 0 {
					// Parse first news
					h.Image = util.GetImageSrc(p)
					h.Category = util.GetText("div.ttprincipal", p)
				}
			} else if s.Is("div") {
				h = Headline{
					Category: util.GetText("div.subn", s),
				}
			} else if s.Is("span") {
				d, _, err := util.CreateDate(util.GetText("", s), " de ")
				if err != nil {
					fmt.Println(err)
				}
				h.Date = &d
			} else if s.Is("h2") || s.Is("h1") {
				h.Title = util.GetText("", s)
				h.NewsPage = s.Find("a").AttrOr("href", "")
			}

			if h.Title != "" && h.NewsPage != "" &&
				!h.Date.IsZero() {
				result = append(result, h)
			}
		})
	})
	err = c.collector.Visit(BaseURL + "/noticias.html")
	return result, err
}

// GetNews TODO
func (h *Headline) GetNews() (*News, error) {
	if h.NewsPage == "" {
		return nil, errors.New("missing news page link")
	}
	return getNews(h.NewsPage)
}

// GetNewsByID TODO
func GetNewsByID(id int) (*News, error) {
	url := fmt.Sprintf("%s/noticia/%d/noticia.html", BaseURL, id)
	return getNews(url)
}

func getNews(url string) (*News, error) {
	var result *News
	var err error

	c := NewClaquete()
	c.collector.OnResponse(func(r *colly.Response) {
		if r.StatusCode >= 200 && r.StatusCode < 301 {
			result = &News{}
		}
	})
	c.collector.OnHTML("body > div.conteudo > div.noticias", func(e *colly.HTMLElement) {
		html, errHTML := e.DOM.Html()
		if errHTML != nil {
			err = errors.Wrapf(errHTML, "couldn't retrive html")
		} else if result != nil {
			result.HTML = strings.TrimSpace(html)

			ds := e.DOM.Find("span:nth-child(1)")
			if ds.Length() != 0 {
				d, _, errDate := util.CreateDate(util.GetText("", ds), " de ")
				if errDate != nil {
					// TODO: proper error handling
					fmt.Println(errDate)
				} else {
					result.Headline.Date = &d
				}
			}

			as := e.DOM.Find("span:nth-child(4)")
			if as.Length() != 0 {
				result.Author = util.GetText("", as)
			}

			hs := e.DOM.Find("h1")
			if hs.Length() != 0 {
				result.Headline.Title = util.GetText("", hs)
			}

			cs := e.DOM.Find("p")
			if cs.Length() != 0 {
				result.Content = util.GetText("", cs)
			}

			slug := util.CreateSlug(result.Headline.Title)
			result.Page = strings.Replace(url, "noticia.html", slug+".html", 1)
			result.Headline.NewsPage = result.Page

			if result.Author == "" {
				err = errors.Wrap(err, "couldn't find author")
			}

			if result.Content == "" {
				err = errors.Wrap(err, "couldn't find content")
			}

			if result.Headline.Title == "" {
				err = errors.Wrap(err, "couldn't find headline")
			}

			if result.Headline.Date.IsZero() {
				err = errors.Wrap(err, "couldn't find date")
			}
		}
	})
	err = c.collector.Visit(url)

	return result, err
}
