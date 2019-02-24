package claquete

import (
	"testing"
	"time"
)

func TestGetCalendar(t *testing.T) {
	calendar, err := GetCalendar()
	if err != nil {
		t.Fatalf("expected no error, but got error: %s", err.Error())
	}

	tt := time.Now()
	if calendar.Month != tt.Month() && calendar.Year != tt.Year() {
		t.Fatalf("expected month/year %d/%d, but got %d/%d",
			tt.Month(), tt.Year(), calendar.Month, calendar.Year)
	}

	calendar, err = calendar.PrevMonth()
	if err != nil {
		t.Fatalf("expected no error, but got error: %s", err.Error())
	}

	tt = tt.AddDate(0, -1, 0)
	if calendar.Month != tt.Month() && calendar.Year != tt.Year() {
		t.Fatalf("expected month/year %d/%d, but got %d/%d",
			tt.Month(), tt.Year(), calendar.Month, calendar.Year)
	}
}
