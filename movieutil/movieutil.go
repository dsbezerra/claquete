package movieutil

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	reOnlyNumbers           = regexp.MustCompile("\\d+")
	reNumbersBetweenSlashes = regexp.MustCompile("\\/(\\d+)\\/")
)

// IDFromURLString alternative IDFromURL that
// receives a string as URL.
func IDFromURLString(u string) (int, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return 0, err
	}
	return IDFromURL(parsedURL)
}

// IDFromURL gets movie ID from these example URLs
// http://www.claquete.com/filmes/filme.php?cf=7785
// http://www.claquete.com/7785/aquaman.html
func IDFromURL(URL *url.URL) (int, error) {
	var ID int
	var err error

	u := URL.String()
	if strings.Contains(u, "/filmes/filme.php?cf=") {
		cf := URL.Query().Get("cf")
		if cf != "" {
			value, err := strconv.Atoi(cf)
			if err != nil {
				err = errors.Wrapf(err, "conversion of cf's value %s to integer failed", cf)
			} else {
				ID = value
			}
		}
	} else {
		r := reNumbersBetweenSlashes.FindStringSubmatch(u)
		if len(r) == 2 {
			value, err := strconv.Atoi(r[1])
			if err != nil {
				err = errors.Wrapf(err, "conversion of %s to integer failed", r[1])
			} else {
				ID = value
			}
		}
	}

	return ID, err
}

// SlugFromURLString alternative SlugFromURL that
// receives a string as URL.
func SlugFromURLString(u string) (string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	return SlugFromURL(parsedURL)
}

// SlugFromURL gets movie slug from this example URL
// http://www.claquete.com/7785/aquaman.html
func SlugFromURL(URL *url.URL) (string, error) {
	var slug string
	var err error

	u := URL.String()
	idx := strings.LastIndex(u, "/")
	if idx > -1 {
		start := idx + 1
		end := len(u)
		if strings.HasSuffix(u, ".html") {
			end -= 5
		}
		slug = u[start:end]
	} else {
		err = fmt.Errorf("couldn't find slug / in url %s", u)
	}

	return slug, err
}

// ParseRuntime converts a runtime text to integer
func ParseRuntime(str string) (int, error) {
	r := reOnlyNumbers.FindString(str)
	v, err := strconv.Atoi(r)
	if err != nil {
		return 0, errors.Wrapf(err, "conversion of %s to integer failed", str)
	}
	return v, err
}

// ParseReleaseDate converts a release date string to time type
func ParseReleaseDate(str, sep string) (*time.Time, error) {
	var result time.Time
	var err error

	parts := strings.Split(str, sep)
	if len(parts) == 3 {
		d, _ := strconv.Atoi(parts[0])
		m, _ := strconv.Atoi(parts[1])
		y, _ := strconv.Atoi(parts[2])
		if d != 0 && m != 0 && y != 0 {
			loc, _ := time.LoadLocation("America/Sao_Paulo")
			result = time.Date(y, time.Month(m), d, 0, 0, 0, 0, loc)
		} else {
			err = fmt.Errorf("couldn't convert text %s to date", str)
		}
	} else {
		err = fmt.Errorf("expected size 3, got %d when splitting text: %s", len(parts), str)
	}

	return &result, err
}
