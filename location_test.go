package claquete

import (
	"testing"
)

func TestGetStates(t *testing.T) {
	states, err := GetStates()
	if err != nil {
		t.Fatalf("expected no error, but got error: %s", err.Error())
	}

	expectedSize := 27
	if expectedSize != len(states) {
		t.Fatalf("expected size %d, got %d", expectedSize, len(states))
	}
}

func TestGetCities(t *testing.T) {
	_, err := GetCities(AC)
	if err != nil {
		t.Fatalf("expected no error, but got error: %s", err.Error())
	}
}

func TestGetCitiesFromState(t *testing.T) {
	states, err := GetStates()
	if err != nil {
		t.Fatalf("expected no error, but got error: %s", err.Error())
	}

	expectedSize := 27
	if expectedSize != len(states) {
		t.Fatalf("expected size %d, got %d", expectedSize, len(states))
	}

	_, err = states[10].GetCities()
	if err != nil {
		t.Fatalf("expected no error, but got error: %s", err.Error())
	}
}

func TestGetTimeZone(t *testing.T) {
	cases := []string{
		"aCre", "America/Rio_Branco",
		"ALAGOAS", "America/Maceio",
		"Amapá", "America/Belem",
		"aMaZoNaS", "America/Manaus",
		"bahia", "America/Bahia",
		"MaraNHão", "America/Fortaleza",
		"MATO grosso DO SUL", "America/Campo_Grande",
		"MATO GROSso", "America/Cuiaba",
		"Pernambuco", "America/Recife",
		"rondônia", "America/Porto_Velho",
		"roRaima", "America/Boa_Vista",
		"toCANtins", "America/Araguaina",
		"MINAS gerais", "America/Sao_Paulo",
	}

	for i := 0; i < len(cases); i += 2 {
		expected := cases[i+1]
		actual := getTimeZone(cases[i])
		if actual != expected {
			t.Fatalf("expected %s, but got %s", expected, actual)
		}
	}
}
