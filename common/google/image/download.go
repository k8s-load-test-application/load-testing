package image

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"sync"
)

type PushHandler func(image []byte) error
type PullHandler func(url string) error

type DownloadImages interface {
	Search(keyword string) ([]Image, error)
	Bulk(keyword string, limit int) (int, error)
}

type downloadImagesImpl struct {
	sizeOfThreadPerFlow []int

	// @NOTE: this callback will be used to push data to another channel
	pushers []PushHandler
	pullers []PullHandler
}

func (self *downloadImagesImpl) Search(keyword string) ([]Image, error) {
	return queryGoogleSearch(buildGoogleSearchQuery(keyword))
}

func (self *downloadImagesImpl) Bulk(keyword string, limit int) (int, error) {
	cnt := 0
	titles := []string{keyword}

	for i := 0; ; i++ {
		nextTitles, size, err := self.doPullAndPush(
			titles[i],
			cnt,
			limit)
		if err != nil {
			return cnt, err
		}
		cnt += size
		titles = append(titles, nextTitles...)
		if cnt >= limit {
			break
		}
	}
	return cnt, nil
}

func (self *downloadImagesImpl) doPullAndPush(
	keyword string,
	current, limit int,
) ([]string, int, error) {
	var wgPull, wgPush sync.WaitGroup

	imageChan := make(chan []byte)
	urlChan := make(chan string)

	// @TODO: split this one to decentralize system
	for id := 0; id < self.sizeOfThreadPerFlow[1]; id++ {
		wgPush.Add(1)

		go func(imageChan chan []byte, wg *sync.WaitGroup) {
			defer wg.Done()

			for image := range imageChan {
				for _, pusher := range self.pushers {
					pusher(image)
				}
			}
		}(imageChan, &wgPush)
	}

	// @TODO: split this one to decentralize system
	for id := 0; id < self.sizeOfThreadPerFlow[0]; id++ {
		wgPull.Add(1)

		go func(id int, wg *sync.WaitGroup) {
			defer wg.Done()

			client := http.DefaultClient
			for url := range urlChan {
				req, _ := http.NewRequest("GET", url, nil)
				resp, err := client.Do(req)
				if err != nil {
					continue
				}

				bytes, err := io.ReadAll(resp.Body)
				if err != nil {
					continue
				}
				for _, puller := range self.pullers {
					puller(url)
				}

				mimetype := http.DetectContentType(bytes)
				if strings.Contains(mimetype, "image") {
					imageChan <- bytes
				} else {
					continue
				}
			}
		}(id, &wgPull)
	}

	titles := make([]string, 0)
	images, err := queryGoogleSearch(buildGoogleSearchQuery(keyword))
	cntFetchedImages := 0
	if err != nil {
		return nil, 0, err
	}

	for _, image := range images {
		cntFetchedImages++
		if cntFetchedImages > limit {
			break
		}

		titles = append(titles, image.Title)
		urlChan <- image.Url
	}

	close(urlChan)
	wgPull.Wait()

	close(imageChan)
	wgPush.Wait()

	return titles, int(cntFetchedImages), nil
}

func parseGoogleSearchResponse(response string) ([]Image, error) {
	var imageJson []interface{}

	scriptStart := strings.LastIndex(response, "AF_initDataCallback")
	if scriptStart == -1 {
		return nil, errors.New("")
	}

	strAfterInitCallback := response[scriptStart:]
	startChar := strings.Index(strAfterInitCallback, "[")
	if startChar == -1 {
		return nil, errors.New("")
	}

	strAfterStartChar := strAfterInitCallback[startChar:]
	endChar := strings.Index(strAfterStartChar, "</script>") - 20
	if endChar <= -1 {
		return nil, errors.New("")
	}

	imageListInStr := strAfterStartChar[:endChar]
	err := json.Unmarshal([]byte(html.UnescapeString(imageListInStr)), &imageJson)
	if err != nil {
		return nil, err
	}

	imageObjects := imageJson[56].([]interface{})[1].([]interface{})[0].([]interface{})[0].([]interface{})[1].([]interface{})[0].([]interface{})
	imageList := make([]Image, 0)

	for _, imageObject := range imageObjects {
		obj := imageObject.([]interface{})[0].([]interface{})[0].(map[string]interface{})["444383007"].([]interface{})[1]
		if obj != nil {
			var image Image
			sourceInfo := obj.([]interface{})[22].(map[string]interface{})["2003"].([]interface{})

			image.Url = obj.([]interface{})[3].([]interface{})[0].(string)
			image.Base = sourceInfo[17].(string)
			image.Title = sourceInfo[3].(string)
			image.Source = sourceInfo[2].(string)

			imageList = append(imageList, image)
		}
	}
	return imageList, nil
}

func queryGoogleSearch(searchQuery string) ([]Image, error) {
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
		return nil, err
	}
	defer resp.Body.Close()

	htmlPage, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseGoogleSearchResponse(string(htmlPage))
}

func buildGoogleSearchQuery(keyword string) string {
	return fmt.Sprintf("https://www.google.com/search?tbm=isch&q=%s",
		strings.Replace(keyword, " ", "+", -1))
}

func NewDownloadImages(
	numberOfThreadPerFlow []int,
	pushers []PushHandler,
	pullers []PullHandler,
) DownloadImages {
	return &downloadImagesImpl{
		sizeOfThreadPerFlow: numberOfThreadPerFlow,
		pushers:             pushers,
		pullers:             pullers,
	}
}
