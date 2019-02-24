package movieutil

import (
	"net/url"
	"testing"
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
