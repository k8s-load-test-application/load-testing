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
	"sync/atomic"
)

type PushHandler func(image []byte) error

type DownloadImages interface {
	Search(keyword string) ([]Image, error)
	Bulk(keyword string) (int, error)
}

type downloadImagesImpl struct {
	sizeOfPullThread int
	sizeOfPushThread int

	// @NOTE: this callback will be used to push data to another channel
	pushers []PushHandler
}

func (self *downloadImagesImpl) Search(keyword string) ([]Image, error) {
	return queryGoogleSearch(buildGoogleSearchQuery(keyword))
}

func (self *downloadImagesImpl) Bulk(keyword string) (int, error) {
	var iwg, ewg sync.WaitGroup

	imageList, err := self.Search(keyword)
	if err != nil {
		return 0, err
	}

	cntFetchedImages := int32(0)

	imageChan := make(chan []byte)
	for id := 0; id < self.sizeOfPushThread; id++ {
		ewg.Add(1)

		go func(imageChan chan []byte, wg *sync.WaitGroup, counter *int32) {
			for image := range imageChan {
				for _, pusher := range self.pushers {
					pusher(image)
				}
				atomic.AddInt32(counter, 1)
			}
		}(imageChan, &ewg, &cntFetchedImages)
	}

	for id := 0; id < self.sizeOfPullThread; id++ {
		iwg.Add(1)

		go func(id int, wg *sync.WaitGroup) {
			defer wg.Done()

			client := http.DefaultClient
			for layer := 0; layer*self.sizeOfPullThread+id < len(imageList); layer++ {
				req, _ := http.NewRequest("GET", imageList[layer*self.sizeOfPullThread+id].Url, nil)
				resp, err := client.Do(req)
				if err != nil {
					continue
				}

				bytes, err := io.ReadAll(resp.Body)
				if err != nil {
					continue
				}

				mimetype := http.DetectContentType(bytes)
				if strings.Contains(mimetype, "image") {
					imageChan <- bytes
				} else {
					continue
				}
			}
		}(id, &iwg)
	}

	iwg.Wait()
	close(imageChan)
	ewg.Wait()

	return int(cntFetchedImages), nil
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
	return fmt.Sprintf("https://www.google.com/search?tbm=isch&q=%s", keyword)
}

func NewDownloadImages(
	numberOfPullThread,
	numberOfPushThread int,
	pushers ...PushHandler,
) DownloadImages {
	return &downloadImagesImpl{
		sizeOfPullThread: numberOfPullThread,
		sizeOfPushThread: numberOfPushThread,
		pushers:          pushers,
	}
}
