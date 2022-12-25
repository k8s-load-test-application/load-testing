package image

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

type DownloadImages interface {
	Bulk(keyword string) (int, error)
}

type downloadImagesImpl struct {
}

func (self *downloadImagesImpl) downloadSingleImage(url string) ([]byte, error) {
	// @TODO: please implement download image from url
	return nil, errors.New("")
}

func (self *downloadImagesImpl) searchImageFromGoogle(keyword string) ([]string, error) {
	loadGoogleSearchResult(keyword)
	return nil, errors.New("")
}

func (self *downloadImagesImpl) Bulk(keyword string) (int, error) {
	// @TODO: please implement bulk download function
	return 0, errors.New("Not implemented")
}

func loadGoogleSearchResult(searchQuery string) (string, error) {
	client := http.DefaultClient
	req, _ := http.NewRequest("GET", searchQuery, nil)

	// @NOTE: No idea why this works, but Google renders the page differently
	//        with this header. Credit to joeclinton1 on Github for this
	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
			"AppleWebKit/537.36 (KHTML, like Gecko) "+
			"Chrome/88.0.4324.104 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	htmlPage, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(htmlPage), nil
}

func buildGoogleSearchQuery(keyword string) string {
	return fmt.Sprintf("https://www.google.com/search?tbm=isch&q=%s", keyword)
}

func New() DownloadImages {
	return &downloadImagesImpl{}
}
