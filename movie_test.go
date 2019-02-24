package claquete

import (
	"strings"
	"testing"
	"time"
)

func TestGetMovie(t *testing.T) {
	movie, err := GetMovie(0)
	if err != errMovieNotFound {
		t.Fatalf("Expected not found, but got %+v", movie)
	}

	movie, err = GetMovie(8427)
	if err != nil {
		t.Fatal(err)
	}

	expectedTitle := "O Retorno de Mary Poppins"
	if movie.Title != expectedTitle {
		t.Fatalf("expected %s, got %s", expectedTitle, movie.Title)
	}

	expectedOriginalTitle := "Mary Poppins Returns"
	if movie.OriginalTitle != expectedOriginalTitle {
		t.Fatalf("expected %s, got %s", expectedOriginalTitle, movie.OriginalTitle)
	}

	expectedRating := -1
	if movie.Rating != expectedRating {
		t.Fatalf("expected %d, got %d", expectedRating, movie.Rating)
	}

	expectedRuntime := 130
	if movie.Runtime != expectedRuntime {
		t.Fatalf("expected %d, got %d", expectedRuntime, movie.Runtime)
	}

	expectedDistributor := "Walt Disney Studios"
	if movie.Distributor != expectedDistributor {
		t.Fatalf("expected %s, got %s", expectedDistributor, movie.Distributor)
	}

	expectedGenres := "Família, Fantasia, Musical"
	if strings.Join(movie.Genres, ", ") != expectedGenres {
		t.Fatalf("expected %s, got %s", expectedGenres, movie.Genres)
	}

	expectedCast := "Emily Blunt, Meryl Streep, Colin Firth, Julie Walters, Ben Whishaw, Emily Mortimer, David Warner, Dick Van Dyke, Angela Lansbury, Lin-Manuel Miranda, Jeremy Swift"
	if strings.Join(movie.Cast, ", ") != expectedCast {
		t.Fatalf("expected %s, got %s", expectedCast, movie.Cast)
	}

	expectedScreenplay := "David Magee"
	if strings.Join(movie.Screenplay, ", ") != expectedScreenplay {
		t.Fatalf("expected %s, got %s", expectedScreenplay, movie.Screenplay)
	}

	expectedProduction := "John DeLuca, Rob Marshall, Marc Platt"
	if strings.Join(movie.Production, ", ") != expectedProduction {
		t.Fatalf("expected %s, got %s", expectedProduction, movie.Production)
	}

	expectedDirection := "Rob Marshall"
	if strings.Join(movie.Direction, ", ") != expectedDirection {
		t.Fatalf("expected %s, got %s", expectedDirection, movie.Direction)
	}

	expectedCountry := "EUA"
	if movie.Country != expectedCountry {
		t.Fatalf("expected %s, got %s", expectedCountry, movie.Country)
	}

	loc, _ := time.LoadLocation("America/Sao_Paulo")
	expectedTime := time.Date(2018, time.December, 20, 0, 0, 0, 0, loc)
	if movie.ReleaseDate == nil || expectedTime.Sub(*movie.ReleaseDate) != 0 {
		t.Fatalf("expected %s, got %s", expectedTime, movie.ReleaseDate)
	}
}
