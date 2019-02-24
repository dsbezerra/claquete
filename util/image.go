package util

import (
	"image" // Added to understand PNG/JPEG formatted images.
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"net/url"
	"time"
)

type (
	Image struct {
		URL           string
		Width, Height int
		Resolution    int
		Format        string
	}
)

// GetImage gets an image config from a given URL
func GetImage(u string) (*Image, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Get(parsedURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	config, format, err := image.DecodeConfig(resp.Body)
	if err != nil {
		return nil, err
	}
	return &Image{
		URL:        parsedURL.String(),
		Width:      config.Width,
		Height:     config.Height,
		Resolution: config.Width * config.Height,
		Format:     format,
	}, nil
}
