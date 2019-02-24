package util

import (
	"fmt"
	"testing"
)

func TestGetImage(t *testing.T) {
	image, err := GetImage("http://www.claquete.com/fotos/filmes/poster/8427_medio.jpg")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(image)
}
