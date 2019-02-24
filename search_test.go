package claquete

import (
	"encoding/json"
	"os"
	"testing"
)

func TestSearch(t *testing.T) {
	results, err := Search("cinemais")
	if err != nil {
		t.Fatal(err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(results)
}

func TestSearchFiltered(t *testing.T) {
	// Retrieves all Avengers movies.
	results, err := SearchFiltered("vingadores", SearchFilterMovie)
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range results.Results {
		if r.Type != SearchTypeMovie {
			t.Fatal("expected all results to be filtered to movie only")
		}
	}

	// This is retrieving results of all three types (Cinema, Movie, News)
	// 21 dec. 2018
	results, err = SearchFiltered("são paulo", SearchFilterMovie|SearchFilterNews)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results.Results {
		if r.Type != SearchTypeMovie && r.Type != SearchTypeCinema {
			t.Fatal("expected all results to be filtered to movie and cinema only")
		}
	}
}
