package claquete

import (
	"testing"
)

func TestGetReleases(t *testing.T) {
	c := NewClaquete()
	_, err := c.GetReleases()
	if err != nil {
		t.Fatalf("expected no error, but got error: %s", err.Error())
	}
}
