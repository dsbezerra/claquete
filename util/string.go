package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// IsUppercase checks if a given token is uppercase
func IsUppercase(tok byte) bool {
	return tok >= 'A' && tok <= 'Z'
}

// EatSpaces eats all space characters found in left and right sides of the given string
func EatSpaces(str *string) {
	size := len(*str)
	if size == 0 {
		return
	}

	i := 0
	le := false // Left eaten
	re := false // Right eaten

	s := 0
	e := size

	for {
		if i == size {
			break
		}

		if (*str)[i] == ' ' && !le {
			s++
		} else {
			le = true
		}

		if (*str)[size-1-i] == ' ' && !re {
			e--
		} else {
			re = true
		}

		if le && re {
			break
		}

		i++
	}

	*str = (*str)[s:e]
}

// BreakByToken break a string in two parts where the tok
// character exists. Otherwise it returns the string
// in the first return value and empty in the second.
func BreakByToken(str string, tok byte) (string, string) {
	if str == "" {
		return "", ""
	}

	i := 0
	size := len(str)

	for i < size {
		if str[i] == tok {
			lhs := str[:i]
			EatSpaces(&lhs)
			rhs := str[i+1:]
			EatSpaces(&rhs)
			return lhs, rhs
		}
		i++
	}

	return str, ""
}

// BreakBySpaces short way to call BreakByToken(str, ' ')
func BreakBySpaces(str string) (string, string) {
	return BreakByToken(str, ' ')
}

func StringToTime(str, sep string) (*time.Time, error) {
	var result time.Time
	var err error

	var d, m, y int
	loc, _ := time.LoadLocation("America/Sao_Paulo")

	parts := strings.Split(str, sep)
	size := len(parts)
	if size > 1 {
		d, _ = strconv.Atoi(parts[0])
		m, _ = strconv.Atoi(parts[1])
		if size == 3 {
			y, _ = strconv.Atoi(parts[2])
		} else {
			y = time.Now().In(loc).Year()
		}
	} else {
		err = fmt.Errorf("expected size 3, got %d when splitting text: %s", len(parts), str)
	}

	if err == nil {
		result = time.Date(y, time.Month(m), d, 0, 0, 0, 0, loc)
	} else {
		err = fmt.Errorf("couldn't convert text %s to date", str)
	}

	return &result, err
}
