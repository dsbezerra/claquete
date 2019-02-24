package claquete

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGetCinema(t *testing.T) {
	cinema, err := GetCinema(656) // Cinemais Montes Claros
	if err != nil {
		t.Fatalf("expected no error, but got error: %s", err.Error())
	}

	if cinema.Name != "Cinemais Montes Claros" {
		t.Fatalf("expected Cinemais Montes Claros, but got error: %s", cinema.Name)
	}
}

func TestGetCinemas(t *testing.T) {
	cinemas, err := GetCinemas(MG, "Montes Claros")
	if err != nil {
		t.Fatalf("expected no error, but got error: %s", err.Error())
	}

	_, err = cinemas[0].GetNowPlaying()
	if err != nil {
		t.Fatalf("expected no error, but got error: %s", err.Error())
	}
}

func TestGetShowtimes(t *testing.T) {
	_, err := GetCinemaShowtimes(58)
	if err != nil {
		t.Fatalf("expected no error, but got error: %s", err.Error())
	}

	s, err := GetCinemaShowtimes(818)
	if err != nil {
		t.Fatalf("expected no error, but got error: %s", err.Error())
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(s)
}
