package movieutil

import (
	"net/url"
	"testing"
	"time"
)

func TestIDFromURL(t *testing.T) {
	u1, _ := url.Parse("http://www.claquete.com/7035/vingadores:-ultimato.html")
	u2, _ := url.Parse("http://www.claquete.com/filmes/filme.php?cf=7035")

	expected := 7035

	ID, _ := IDFromURL(u1)
	if expected != ID {
		t.Fatalf("expected %d, got %d", expected, ID)
	}

	ID, _ = IDFromURL(u2)
	if expected != ID {
		t.Fatalf("expected %d, got %d", expected, ID)
	}
}

func TestSlugFromURL(t *testing.T) {
	_, err := SlugFromURLString("invalid url")
	if err == nil {
		t.Fatal("expected error")
	}

	u := "http://www.claquete.com/7785/aquaman.html"

	expected := "aquaman"
	actual, err := SlugFromURLString(u)
	if err != nil {
		t.Fatalf("expected no error, but got error: %s", err.Error())
	}

	if expected != actual {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
}

func TestParseRuntime(t *testing.T) {
	actual, err := ParseRuntime("90")
	if err != nil {
		t.Fatalf("expected no error, but got error: %s", err.Error())
	}

	expected := 90
	if actual != expected {
		t.Fatalf("expected %d, got %d", expected, actual)
	}

	_, err = ParseRuntime("definitely not a runtime")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseReleaseDate(t *testing.T) {
	actual, err := ParseReleaseDate("07/03/2019", "/")
	if err != nil {
		t.Fatalf("expected no error, but got error: %s", err.Error())
	}

	expected := time.Date(2019, time.March, 7, 0, 0, 0, 0, time.UTC)

	eYear, eMonth, eDay := expected.Date()
	if eYear != actual.Year() || eMonth != actual.Month() || eDay != actual.Day() {
		t.Fatalf("expected %s, but got %s", expected, actual)
	}
}
