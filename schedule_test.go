package claquete

import (
	"testing"
)

func TestGetSchedule(t *testing.T) {
	sched, err := GetSchedule(656) // Cinemais Montes Claros
	if err != nil {
		t.Fatalf("expected no error, but got error: %s", err.Error())
	}

	if sched.Cinema.Name != "Cinemais Montes Claros" {
		t.Fatalf("expected Cinemais Montes Claros, got %s", sched.Cinema.Name)
	}
}
