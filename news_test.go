package claquete

import (
	"testing"
)

func TestGetHeadlines(t *testing.T) {
	results, err := GetHeadlines()
	if err != nil {
		t.Fatal("expected no error")
	}

	if len(results) == 0 {
		t.Fatal("expected result size to be different than zero")
	}

	// First headline contains image
	first := results[0]
	if first.Image == "" {
		t.Fatal("expected image url in first headline but got nothing")
	}
}

func TestGetNewsByID(t *testing.T) {
	result, err := GetNewsByID(10425)
	expected := News{
		Author: "Fernanda Mendes",
		Headline: Headline{
			Title: "Paris Filmes fecha contrato para filme sobre Ney Matogrosso",
		},
		Page: "http://claquete.com.br/noticia/10425/paris-filmes-fecha-contrato-para-filme-sobre-ney-matogrosso.html",
	}

	if err != nil {
		t.Fatal("expected no error")
	}

	if result.Headline.Title != expected.Headline.Title {
		t.Fatalf("expected title %s, but got %s", expected.Headline.Title, result.Headline.Title)
	}

	if result.Author != expected.Author {
		t.Fatalf("expected author %s, but got %s", expected.Author, result.Author)
	}

	if result.Page != expected.Page {
		t.Fatalf("expected page %s, but got %s", expected.Page, result.Page)
	}
}
