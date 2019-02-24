package util

import (
	"testing"
)

func TestEatSpaces(t *testing.T) {
	expected := "testing"
	str := "         testing      "
	EatSpaces(&str)
	if str != expected {
		t.Fatalf("expected %s, got %s", expected, str)
	}
}

func TestBreakBySpaces(t *testing.T) {
	str := "testing testing2"

	expectedl := "testing"
	expectedr := "testing2"

	lhs, rhs := BreakBySpaces(str)
	if lhs != expectedl {
		t.Fatalf("expected %s as left-hand side, got %s", expectedl, lhs)
	}

	if rhs != expectedr {
		t.Fatalf("expected %s as right-hand side, got %s", expectedr, rhs)
	}
}

func TestBreakByToken(t *testing.T) {
	str := "testing : testing2"

	expectedl := "testing"
	expectedr := "testing2"

	lhs, rhs := BreakByToken(str, ':')
	if lhs != expectedl {
		t.Fatalf("expected %s as left-hand side, got %s", expectedl, lhs)
	}

	if rhs != expectedr {
		t.Fatalf("expected %s as right-hand side, got %s", expectedr, rhs)
	}
}
