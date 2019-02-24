package util

import "testing"

func TestCreateSlug(t *testing.T) {
	title := "Vingadores: Ultimato"
	expected := "vingadores:-ultimato"

	slug := CreateSlug(title)
	if expected != slug {
		t.Fatalf("expected %s, got %s", expected, slug)
	}

	title = "Detetives do Prédio Azul 2 - O Mistério Italiano"
	expected = "detetives-do-predio-azul-2--o-misterio-italiano"

	slug = CreateSlug(title)
	if expected != slug {
		t.Fatalf("expected %s, got %s", expected, slug)
	}
}
