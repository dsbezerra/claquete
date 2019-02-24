package util

import (
	"bytes"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	errInvalidOption = errors.New("invalid option")
)

// UserAgents is a list of user agents used in HTTP requests
var UserAgents = [...]string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:66.0) Gecko/20100101 Firefox/66.0",
	"Mozilla/5.0 (X11; Linux i686; rv:64.0) Gecko/20100101 Firefox/64.0",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:64.0) Gecko/20100101 Firefox/64.0",
	"Mozilla/5.0 (X11; Linux i586; rv:63.0) Gecko/20100101 Firefox/63.0",
	"Mozilla/5.0 (Windows NT 6.2; WOW64; rv:63.0) Gecko/20100101 Firefox/63.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.36 (KHTML like Gecko) Chrome/44.0.2403.155 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML like Gecko) Chrome/51.0.2704.79 Safari/537.36 Edge/14.14931",
	"Opera/9.80 (X11; Linux i686; Ubuntu/14.10) Presto/2.12.388 Version/12.16",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_3) AppleWebKit/537.75.14 (KHTML, like Gecko) Version/7.0.3 Safari/7046A194A",
}

// RandomUserAgent retrieves a random user agent
func RandomUserAgent() string {
	result := ""

	// Using current time nanosecond as seed
	seed := time.Now().Nanosecond()

	// Seed the random
	rand.Seed(int64(seed))

	// Get random user-agent
	size := len(UserAgents)
	result = UserAgents[rand.Int31n(int32(size))]

	return result
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

// CreateSlug from string by replacing all spaces with -
// and uppercase characters with lower ones.
func CreateSlug(title string) string {
	if title == "" {
		return ""
	}

	var result bytes.Buffer

	b := make([]byte, len(title))
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	_, _, e := t.Transform(b, []byte(title), true)
	if e != nil {
		fmt.Println(e)
		return ""
	}

	stripped := string(b)

	i := 0
	size := len(stripped)
	for i < size {
		b := stripped[i]
		if unicode.IsUpper(rune(b)) {
			b += 32
		} else if unicode.IsSpace(rune(b)) {
			b = 0
			if i+1 < size {
				next := stripped[i+1]
				if next != '-' {
					b = '-'
				}
			}
		}
		if b != 0 {
			result.WriteByte(b)
		}
		i++
	}

	return result.String()
}

func CreateDate(str, sep string) (time.Time, bool, error) {
	var result time.Time
	var ok bool
	var err error

	parts := strings.Split(str, sep)
	if len(parts) == 3 {
		ok = true

		d, err := strconv.Atoi(parts[0])
		if err != nil {
			ok = false
		}

		m := GetMonth(parts[1])
		if m < time.January || m > time.December {
			ok = false
		}

		y, err := strconv.Atoi(parts[2])
		if err != nil {
			ok = false
		}

		if ok {
			loc, _ := time.LoadLocation("America/Sao_Paulo")
			result = time.Date(y, m, d, 0, 0, 0, 0, loc)
		} else {
			err = fmt.Errorf("couldn't create date from text: %s", str)
		}
	} else {
		err = fmt.Errorf("expected size 3, got: %d", len(parts))
	}

	return result, ok, err
}

func GetText(sel string, s *goquery.Selection) string {
	if sel != "" {
		s = s.Find(sel)
	}
	return strings.TrimSpace(s.Text())
}

func GetImageSrc(s *goquery.Selection) string {
	return s.Find("img").First().AttrOr("src", "")
}

// GetMonth returns the time.Month correspond to the brazilian month name.
func GetMonth(str string) time.Month {
	var result time.Month

	str = strings.ToLower(str)
	switch str {
	case "janeiro":
		result = time.January
	case "fevereiro":
		result = time.February
	case "março":
		result = time.March
	case "abril":
		result = time.April
	case "maio":
		result = time.May
	case "junho":
		result = time.June
	case "julho":
		result = time.July
	case "agosto":
		result = time.August
	case "setembro":
		result = time.September
	case "outubro":
		result = time.October
	case "novembro":
		result = time.November
	case "dezembro":
		result = time.December
	}

	return result
}
