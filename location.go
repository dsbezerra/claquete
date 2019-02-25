package claquete

import (
	"strings"

	"github.com/gocolly/colly"
)

const (
	// FederativeUnit (FU) is UF

	// AC is Acre's FU
	AC = "AC"
	// AL is Alagoas's FU
	AL = "AL"
	// AM is Amazona's FU
	AM = "AM"
	// AP is Amapa's FU
	AP = "AP"
	// BA is Bahia's FU
	BA = "BA"
	// CE is Ceará's FU
	CE = "CE"
	// DF is Distrito Federal's FU
	DF = "DF"
	// ES is Espírito Santo's FU
	ES = "ES"
	// GO is Goiás's FU
	GO = "GO"
	// MA is Maranhão's FU
	MA = "MA"
	// MG is Minas Gerais's FU
	MG = "MG"
	// MS is Mato Grosso do Sul's FU
	MS = "MS"
	// MT is Mato Grosso's FU
	MT = "MT"
	// PA is Pará's FU
	PA = "PA"
	// PB is Paraíba's FU
	PB = "PB"
	// PE is Pernambuco's FU
	PE = "PE"
	// PI is Piauí's FU
	PI = "PI"
	// PR is Paraná's FU
	PR = "PR"
	// RJ is Rio de Janeiro's FU
	RJ = "RJ"
	// RN is Rio Grande do Norte's FU
	RN = "RN"
	// RO is Rondônia's FU
	RO = "RO"
	// RR is Roraima's FU
	RR = "RR"
	// RS is Rio Grande do Sul's FU
	RS = "RS"
	// SC is Santa Catarina's FU
	SC = "SC"
	// SE is Sergipe's FU
	SE = "SE"
	// SP is São Paulo's FU
	SP = "SP"
	// TO is Tocantins's FU
	TO = "TO"
)

const (
	Acre            = "Acre"
	Alagoas         = "Alagoas"
	Amazonas        = "Amazonas"
	Amapa           = "Amapá"
	Bahia           = "Bahia"
	Ceara           = "Ceará"
	DistritoFederal = "Distrito Federal"
	EspiritoSanto   = "Espírito Santo"
	Goias           = "Goiás"
	Maranhao        = "Maranhão"
	MinasGerais     = "Minas Gerais"
	MatoGrossoSul   = "Mato Grosso do Sul"
	MatoGrosso      = "Mato Grosso"
	Para            = "Pará"
	Paraiba         = "Paraíba"
	Pernambuco      = "Pernambuco"
	Piaui           = "Piauí"
	Parana          = "Paraná"
	RioJaneiro      = "Rio de Janeiro"
	RioGrandeNorte  = "Rio Grande do Norte"
	Rondonia        = "Rondônia"
	Roraima         = "Roraima"
	RioGrandeSul    = "Rio Grande do Sul"
	SantaCatarina   = "Santa Catarina"
	Sergipe         = "Sergipe"
	SaoPaulo        = "São Paulo"
	Tocantins       = "Tocantins"
)

var (
	// FederativeUnits is used only to check if input
	// is valid.
	FederativeUnits = []string{
		AC, AL, AM, AP, BA, CE, DF, ES, GO, MA, MG, MS,
		MT, PA, PB, PE, PI, PR, RJ, RN, RO, RR, RS, SC,
		SE, SP, TO,
	}
)

type (
	// City is a very simple representation of a city.
	City struct {
		c     *Claquete
		State State  `json:"state"`
		Name  string `json:"name"`
	}

	// State is a very simple representation of a state.
	State struct {
		c    *Claquete
		FU   string `json:"fu"`
		Name string `json:"name"`
	}
)

// GetStates retrieves list of states.
func GetStates() ([]State, error) {
	var result []State
	var err error

	c := NewClaquete()
	c.collector.OnHTML("#selUf > option", func(e *colly.HTMLElement) {
		value := e.Attr("value")
		if isFederativeUnitValid(value) {
			result = append(result, State{
				c:    c,
				FU:   value,
				Name: e.Text,
			})
		}
	})
	err = c.collector.Visit(BaseURL)
	return result, err
}

// GetCities retrieve list of cities.
func (s *State) GetCities() ([]City, error) {
	var result []City
	var err error

	gotCookies := false
	s.c.collector.AllowURLRevisit = true
	s.c.collector.OnResponse(func(r *colly.Response) {
		// Now with cookies we visit the index page
		// to get the list of cities.
		if !gotCookies {
			gotCookies = true
			s.c.collector.Visit(BaseURL)
		}
	})

	s.c.collector.OnHTML("#cidade > option", func(e *colly.HTMLElement) {
		value := e.Attr("value")
		if value != "0" {
			result = append(result, City{
				c:    s.c,
				Name: e.Text,
			})
		}
	})

	// POST will set cookies.
	err = s.c.collector.Post(BaseAJAX+"escolherEstados.php", map[string]string{
		"UF": string(s.FU),
	})

	s.c.collector.AllowURLRevisit = false

	return result, err
}

// GetNowPlaying retrieves now playing movies for City.
// To get additional movie metadata use the GetMovie(int)
// function passing the retrieved movie id.
func (c *City) GetNowPlaying() ([]Movie, error) {
	params := map[string]string{"cidade": c.Name}
	return getNowPlayingList(c.c.collector, "escolherFilme_cidade.php", params)
}

// GetCities for current claquete's federative unit
func (c *Claquete) GetCities() ([]City, error) {
	s := &State{c: c, FU: c.fu, Name: getStateName(c.fu)}
	return s.GetCities()
}

// GetCities for specific Federative Unit independent of
// Claquete instance
func GetCities(fu string) ([]City, error) {
	c := NewClaquete(FederativeUnit(fu))
	return c.GetCities()
}

func getStateName(s string) string {
	switch s {
	case AC:
		return Acre
	case AL:
		return Alagoas
	case AM:
		return Amazonas
	case AP:
		return Amapa
	case BA:
		return Bahia
	case CE:
		return Ceara
	case DF:
		return DistritoFederal
	case ES:
		return EspiritoSanto
	case GO:
		return Goias
	case MA:
		return Maranhao
	case MG:
		return MinasGerais
	case MS:
		return MatoGrossoSul
	case MT:
		return MatoGrosso
	case PA:
		return Para
	case PB:
		return Paraiba
	case PE:
		return Pernambuco
	case PI:
		return Piaui
	case PR:
		return Parana
	case RJ:
		return RioJaneiro
	case RN:
		return RioGrandeNorte
	case RO:
		return Rondonia
	case RR:
		return Roraima
	case RS:
		return RioGrandeSul
	case SC:
		return SantaCatarina
	case SE:
		return Sergipe
	case SP:
		return SaoPaulo
	case TO:
		return Tocantins
	default:
		return ""
	}
}
func getTimeZone(s string) string {
	result := "America/Sao_Paulo"

	if strings.EqualFold(s, Acre) {
		result = "America/Rio_Branco"
	} else if strings.EqualFold(s, Alagoas) || strings.EqualFold(s, Sergipe) {
		result = "America/Maceio"
		// NOTE: or America/Santarem for west of Para, refactor if necessary
	} else if strings.EqualFold(s, Amapa) || strings.EqualFold(s, Para) {
		result = "America/Belem"
		// NOTE: or America/Eirunepe for west, refactor if necessary
	} else if strings.EqualFold(s, Amazonas) {
		result = "America/Manaus"
	} else if strings.EqualFold(s, Bahia) {
		result = "America/Bahia"
	} else if strings.EqualFold(s, Maranhao) || strings.EqualFold(s, Piaui) ||
		strings.EqualFold(s, Ceara) || strings.EqualFold(s, RioGrandeNorte) ||
		strings.EqualFold(s, Paraiba) {
		result = "America/Fortaleza"
	} else if strings.EqualFold(s, MatoGrossoSul) {
		result = "America/Campo_Grande"
	} else if strings.EqualFold(s, MatoGrosso) {
		result = "America/Cuiaba"
	} else if strings.EqualFold(s, Pernambuco) {
		result = "America/Recife"
	} else if strings.EqualFold(s, Rondonia) {
		result = "America/Porto_Velho"
	} else if strings.EqualFold(s, Roraima) {
		result = "America/Boa_Vista"
	} else if strings.EqualFold(s, Tocantins) {
		result = "America/Araguaina"
	}

	return result
}

func isFederativeUnitValid(fu string) bool {
	for _, f := range FederativeUnits {
		if f == fu {
			return true
		}
	}
	return false
}
