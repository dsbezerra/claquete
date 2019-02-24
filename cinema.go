package claquete

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/dsbezerra/claqueteapi/movieutil"
	"github.com/dsbezerra/claqueteapi/util"
	"github.com/gocolly/colly"
	"github.com/pkg/errors"
)

// NoteType TODO
type NoteType string

const (
	// NoteOnlyDayX TODO
	NoteOnlyDayX = "only"
	// NoteExceptDayX TODO
	NoteExceptDayX = "except"

	// VersionDubbed TODO
	VersionDubbed = "dubbed"
	// VersionNational TODO
	VersionNational = "national"
	// VersionSubtitled TODO
	VersionSubtitled = "subtitled"

	// Format2D TODO
	Format2D = "2D"
	// Format3D TODO
	Format3D = "3D"
	// Format4DX TODO
	Format4DX = "4DX"
)

var (
	// PeriodExp TODO
	PeriodExp = regexp.MustCompile("\\((\\d{2}\\/\\d{2})\\)")
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

	// Period TODO
	Period struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	}

	// Showtime TODO
	Showtime struct {
		CinemaID   int        `json:"cinema_id"`
		MovieID    int        `json:"movie_id"`
		MovieTitle string     `json:"movie_title"`
		Format     string     `json:"format"`
		Version    string     `json:"version"`
		StartTime  *time.Time `json:"opening_time"`
		Room       int        `json:"room"`
		Weekday    int        `json:"weekday"`
		Period     Period     `json:"period"`
		VIP        bool       `json:"vip"`
		XD         bool       `json:"xd"`
		IMAX       bool       `json:"imax"`
	}

	// NoteMap TODO
	NoteMap struct {
		sync.RWMutex
		m map[string]Note
	}

	// Note TODO
	Note struct {
		Type NoteType
		Days []time.Time
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
		addressLine := strings.TrimSpace(e.DOM.Find("div > div > span").Text())
		if addressLine != "" {
			addressLine = strings.Replace(addressLine, "\n", "", -1)
			addressLine = strings.Replace(addressLine, "(mapa)", "", -1)
		}
		result = &Cinema{
			c:           c,
			ID:          id,
			Name:        strings.TrimSpace(e.DOM.Find("div > div > div.ttcine > h2").Text()),
			AddressLine: strings.TrimSpace(addressLine),
		}
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

// GetCinemaShowtimes TODO
func GetCinemaShowtimes(id int) ([]Showtime, error) {
	if id < 0 {
		return nil, errors.New("invalid ID")
	}

	idStr := strconv.Itoa(id)
	if idStr == "" {
		return nil, fmt.Errorf("couldn't create ID from integer %d", id)
	}

	var period Period
	var result []Showtime
	var err error

	exceptOnlyMap := &NoteMap{m: make(map[string]Note)}

	container := fmt.Sprintf("body > div.conteudo > div.progrb > div.cinema%d", id)

	c := NewClaquete()
	c.collector.OnHTML(container, func(e *colly.HTMLElement) {
		periods := e.DOM.Find("h4 > strong")
		if periods.Length() == 2 {
			st := util.GetText("", periods.First())
			et := util.GetText("", periods.Last())
			start, _ := util.StringToTime(st, "/")
			end, _ := util.StringToTime(et, "/")

			if start != nil && end != nil {
				period.Start = *start
				period.End = *end
			}
		}
		e.DOM.Find("div.filme").Each(func(i int, s *goquery.Selection) {
			a := s.Find("h2 > a")
			t := strings.TrimSpace(a.Text())
			if t == "" {
				// TODO: error
				return
			}

			var idMovie, room int
			var title string

			idMovie, err := movieutil.IDFromURLString(a.AttrOr("href", ""))
			if err != nil {
				// TODO: error
				return
			}

			title = t

			fillExceptOnlyMap(exceptOnlyMap, s.Find("span.hleter"))

			remainder := strings.TrimSpace(s.Find("h2.salas").Text())
			if remainder != "" {
				lhs, rhs := util.BreakBySpaces(remainder)
				if lhs == "Sala" {
					num, rhs := util.BreakBySpaces(rhs)
					if num != "" {
						value, err := strconv.Atoi(num)
						if err != nil {
							fmt.Printf("couldn't convert room number %s to integer", num)
							return
						}
						room = value
						remainder = rhs
					}
				}
			}

			if room == 0 {
				// Skip
				return
			}

			showtime := Showtime{
				CinemaID:   id,
				MovieID:    idMovie,
				MovieTitle: title,
				Room:       room,
				Version:    Format2D,
				Period:     period,
			}

			s.Find("div.icons div").Each(func(i int, s *goquery.Selection) {
				hint := s.AttrOr("data-hint", "")
				if hint != "" {
					hint = strings.ToLower(hint)
					if strings.Contains(hint, "dub") {
						showtime.Version = VersionDubbed
					} else if strings.Contains(hint, "leg") {
						showtime.Version = VersionSubtitled
					} else if strings.Contains(hint, "nac") {
						showtime.Version = VersionNational
					} else if strings.Contains(hint, "3d") {
						showtime.Format = Format3D
					} else if strings.Contains(hint, "4dx") {
						showtime.Format = Format4DX
					} else if strings.Contains(hint, "vip") {
						showtime.VIP = true
					} else if strings.Contains(hint, "xd") {
						showtime.XD = true
					} else if strings.Contains(hint, "imax") {
						showtime.IMAX = true
					}
				}
			})

			var ot string
			for remainder != "" {
				ot, remainder = util.BreakByToken(remainder, ',')
				if ot == "" {
					break
				}

				var letter string
				var size = len(ot)
				// If time has a letter as last character remove it and update time string
				if size > 0 && unicode.IsLetter(rune(ot[size-1])) {
					letter = ot[size-1:]
					ot = ot[0 : size-1]
				}

				h, m := util.BreakByToken(ot, 'h')
				if h == "" || m == "" {
					// TODO: Proper error handling
					return
				}

				if err != nil {
					// TODO: Proper error handling
					return
				}

				hours, err := strconv.Atoi(h)
				if err != nil {
					// TODO: Proper error handling
					return
				}

				minutes, err := strconv.Atoi(m)
				if err != nil {
					// TODO: Proper error handling
					return
				}

				if letter != "" {
					exceptOnlyMap.RLock()
					n, ok := exceptOnlyMap.m[letter]
					exceptOnlyMap.RUnlock()
					if ok {
						s := showtime
						if n.Type == NoteOnlyDayX {
							for _, day := range n.Days {
								y, m, d := day.Date()
								st := time.Date(y, m, d, hours, minutes, 0, 0, day.Location())
								s.StartTime = &st
								result = append(result, s)
							}
						} else {
							for d := 0; d < 7; d++ {
								nd := period.Start.AddDate(0, 0, d)
								add := true
								for _, day := range n.Days {
									// NOTE: Only day and month are enough here because we use
									// AddDate above which guarantees the correctness of time
									// when we deal with dates near New Year's Eve
									_, m, d := day.Date()
									if nd.Day() == d && nd.Month() == m {
										add = false
										break
									}
								}

								if add {
									s.StartTime = setTime(&nd, hours, minutes)
									result = append(result, s)
								}
							}
						}
					} else {
						// TODO: Proper error handling
					}
				} else {
					// Build showtimes for the whole week
					for d := 0; d < 7; d++ {
						nd := period.Start.AddDate(0, 0, d)
						showtime.StartTime = setTime(&nd, hours, minutes)
						result = append(result, showtime)
					}
				}
			}
		})
	})

	// cinema-%s.html is optional, the request is successful with any string
	// followed by .html
	u := fmt.Sprintf("%s/programacao/%s/cinema-%s.html", BaseURL, idStr, idStr)
	err = c.collector.Visit(u)
	return result, err
}

// GetNowPlaying retrieves now playing movies for Cinema.
// To get additional movie metadata use the GetMovie(int)
// function passing the retrieved movie id.
func (c *Cinema) GetNowPlaying() ([]Movie, error) {
	params := map[string]string{"cinema": strconv.Itoa(c.ID)}
	return getNowPlayingList(c.c.collector, "escolherFilme.php", params)
}

func (c *Claquete) GetCinemas() ([]Cinema, error) {
	return getCinemas(c)
}

func getCinemas(c *Claquete) ([]Cinema, error) {
	var result []Cinema
	var err error

	if c.CityName == "" {
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
	err = c.collector.Post(u, map[string]string{"cidade": c.CityName})
	return result, err
}

// GetCinemas retrieve cinema list
func GetCinemas(fu, city string) ([]Cinema, error) {
	return getCinemas(NewClaquete(
		FederativeUnit(fu),
		CityName(city),
	))
}

func fillExceptOnlyMap(nm *NoteMap, s *goquery.Selection) {
	s.Each(func(i int, ss *goquery.Selection) {
		n := strings.TrimSpace(ss.Text())
		nm.RLock()
		_, ok := nm.m[n]
		nm.RUnlock()
		if !ok {
			hint := ss.Parent().AttrOr("data-hint", "")
			if hint == "" {
				fmt.Printf("couldn't find hint for note %s\n", n)
			} else {

				var result Note
				hint = strings.TrimSpace(strings.ToLower(hint))
				t, remainder := util.BreakBySpaces(hint)
				if strings.Contains(t, "somente") {
					result.Type = NoteOnlyDayX
				} else if strings.Contains(t, "exceto") {
					result.Type = NoteExceptDayX
				} else {
					fmt.Printf("unknown value '%s' for type\n", t)
				}

				if result.Type != "" {
					// Parse day
					remainder = strings.Replace(remainder, ".", " ", -1)

					var day, date string
					// Get the day(s) now
					for remainder != "" {
						day, remainder = util.BreakBySpaces(remainder)
						if day == "" {
							break
						}

						if remainder == "" {
							break
						}

						date, remainder = util.BreakBySpaces(remainder)
						if date == "" {
							break
						}

						res := PeriodExp.FindStringSubmatch(date)
						if len(res) == 2 {
							d, err := util.StringToTime(res[1], "/")
							if err != nil {
								fmt.Printf("Failed to create date from text '%s'\n", res[1])
							} else {
								result.Days = append(result.Days, *d)
							}
						}
					}

					nm.Lock()
					nm.m[n] = result
					nm.Unlock()
				}
			}
		}
	})
}

func setTime(t *time.Time, hours, minutes int) *time.Time {
	result := *t
	result = result.
		Add(time.Duration(hours) * time.Hour).
		Add(time.Duration(minutes) * time.Minute)
	return &result
}

func (c *Cinema) GetShowtimes() ([]Showtime, error) {
	return GetCinemaShowtimes(c.ID)
}
