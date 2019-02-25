package claquete

import (
	"errors"
	"fmt"
	"log"
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
	// Schedule ...
	Schedule struct {
		Cinema   *Cinema   `json:"cinema"`
		Period   *Period   `json:"period"`
		Sessions []Session `json:"sessions"`
		loc      *time.Location
		dis      *NoteMap
	}

	// Period TODO
	Period struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	}

	// Session TODO
	Session struct {
		CinemaID   int        `json:"cinema_id"`
		MovieID    int        `json:"movie_id"`
		MovieTitle string     `json:"movie_title"`
		Format     string     `json:"format"`
		Version    string     `json:"version"`
		StartTime  *time.Time `json:"opening_time"`
		Room       int        `json:"room"`
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
		Type NoteType    `json:"type"`
		Days []time.Time `json:"days"`
	}
)

// GetSchedule TODO
func GetSchedule(cinema int) (*Schedule, error) {
	if cinema < 0 {
		return nil, errors.New("invalid ID")
	}

	idStr := strconv.Itoa(cinema)
	if idStr == "" {
		return nil, fmt.Errorf("couldn't create ID from integer %d", cinema)
	}

	var result *Schedule
	var err error

	c := NewClaquete()

	c.collector.OnHTML("body > div.conteudo >div.progrb", func(e *colly.HTMLElement) {
		schedule, err := parseSchedule(cinema, e.DOM)
		if err != nil {
			return
		}
		result = schedule
	})

	// cinema-%s.html is optional, the request is successful with any string
	// followed by .html
	u := fmt.Sprintf("%s/programacao/%s/cinema-%s.html", BaseURL, idStr, idStr)
	err = c.collector.Visit(u)
	return result, err
}

func (sched *Schedule) fillDisclaimer(s *goquery.Selection) {
	s.Each(func(i int, ss *goquery.Selection) {
		n := strings.TrimSpace(ss.Text())
		sched.dis.RLock()
		_, ok := sched.dis.m[n]
		sched.dis.RUnlock()
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
							d, err := util.StringToTime(res[1], "/", sched.loc)
							if err != nil {
								fmt.Printf("Failed to create date from text '%s'\n", res[1])
							} else {
								result.Days = append(result.Days, *d)
							}
						}
					}

					sched.dis.Lock()
					sched.dis.m[n] = result
					sched.dis.Unlock()
				}
			}
		}
	})
}

func parseSchedule(c int, s *goquery.Selection) (*Schedule, error) {
	cinema, err := parseCinema(s)
	if err != nil {
		// TODO: proper error handling
		fmt.Println(err)
		return nil, err
	}
	cinema.ID = c

	result := &Schedule{}

	var disclaimer = &NoteMap{
		m: make(map[string]Note),
	}
	var loc *time.Location
	// Try to retrieve time zone for cinema
	{
		e := s.Find("div:nth-child(1) > p")
		t := strings.TrimSpace(strings.Replace(e.Text(), "por cinemas em", "", -1))
		if t != "" { // Expected state name
			tz := getTimeZone(t)
			cinema.TimeZone = tz

			loc, _ = time.LoadLocation(tz)
			result.loc = loc
			result.dis = disclaimer
		}
	}
	result.Cinema = cinema

	s.Find("div").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if class := s.AttrOr("class", ""); strings.HasPrefix(class, "cinema") {
			var err error

			dates := s.Find("h4 > strong")
			if dates.Length() == 2 {
				start, _ := util.StringToTime(util.GetText("", dates.First()), "/", loc)
				end, _ := util.StringToTime(util.GetText("", dates.Last()), "/", loc)
				if start != nil && end != nil {
					result.Period = &Period{
						Start: *start,
						End:   *end,
					}
				} else {
					err = errors.New("couldn't retrieve schedule period")
				}
			} else {
				err = errors.New("unexpected DOM structure")
			}

			if err != nil {
				// TODO: proper error handling
				log.Fatal(err)
			}
			result.fillDisclaimer(s.Find("span.hleter"))
			s.Find("div.filme").Each(func(i int, s *goquery.Selection) {
				sessions, err := parseSessions(s, result)
				if err != nil {
					// TODO: proper error handling
					log.Fatal(err)
				}

				result.Sessions = append(result.Sessions, sessions...)
			})
		}

		return true
	})

	return result, err
}

func parseSessions(s *goquery.Selection, sched *Schedule) ([]Session, error) {
	var result []Session

	a := s.Find("h2 > a")
	t := strings.TrimSpace(a.Text())
	if t == "" {
		return nil, errors.New("couldn't find movie title")
	}

	var idMovie, room int
	var title string

	idMovie, err := movieutil.IDFromURLString(a.AttrOr("href", ""))
	if err != nil {
		// TODO: error
		return nil, err
	}

	title = t

	remainder := strings.TrimSpace(s.Find("h2.salas").Text())
	if remainder != "" {
		lhs, rhs := util.BreakBySpaces(remainder)
		if lhs == "Sala" {
			num, rhs := util.BreakBySpaces(rhs)
			if num != "" {
				value, err := strconv.Atoi(num)
				if err != nil {
					fmt.Printf("couldn't convert room number %s to integer", num)
					return nil, err
				}
				room = value
				remainder = rhs
			}
		}
	}

	if room == 0 {
		// Skip
		return nil, err
	}

	session := Session{
		CinemaID:   sched.Cinema.ID,
		MovieID:    idMovie,
		MovieTitle: title,
		Room:       room,
		Format:     Format2D,
	}

	s.Find("div.icons div").Each(func(i int, s *goquery.Selection) {
		hint := s.AttrOr("data-hint", "")
		if hint != "" {
			hint = strings.ToLower(hint)
			if strings.Contains(hint, "dub") {
				session.Version = VersionDubbed
			} else if strings.Contains(hint, "leg") {
				session.Version = VersionSubtitled
			} else if strings.Contains(hint, "nac") {
				session.Version = VersionNational
			} else if strings.Contains(hint, "3d") {
				session.Format = Format3D
			} else if strings.Contains(hint, "4dx") {
				session.Format = Format4DX
			} else if strings.Contains(hint, "vip") {
				session.VIP = true
			} else if strings.Contains(hint, "xd") {
				session.XD = true
			} else if strings.Contains(hint, "imax") {
				session.IMAX = true
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
			err = errors.New("couldn't find time")
		}

		if err != nil {
			// TODO: Proper error handling
			return nil, err
		}

		hours, err := strconv.Atoi(h)
		if err != nil {
			return nil, err
		}

		minutes, err := strconv.Atoi(m)
		if err != nil {
			return nil, err
		}

		if letter != "" {
			sched.dis.RLock()
			n, ok := sched.dis.m[letter]
			sched.dis.RUnlock()
			if ok {
				s := session
				if n.Type == NoteOnlyDayX {
					for _, day := range n.Days {
						y, m, d := day.Date()
						st := time.Date(y, m, d, hours, minutes, 0, 0, sched.loc)
						s.StartTime = &st
						result = append(result, s)
					}
				} else {
					for d := 0; d < 7; d++ {
						nd := sched.Period.Start.AddDate(0, 0, d)
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
			// Build sessions for the whole week
			for d := 0; d < 7; d++ {
				nd := sched.Period.Start.AddDate(0, 0, d)
				session.StartTime = setTime(&nd, hours, minutes)
				result = append(result, session)
			}
		}
	}

	return result, err
}
