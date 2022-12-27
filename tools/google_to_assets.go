package main

import (
	"flag"
	"fmt"

	gg "github.com/k8s-load-test-application/load-testing/common/google/image"
)

var keyword string
var thread int
var limit int

func pushImageToAssets(image []byte) error {
	return nil
}

func pullImageToAssets(url string) error {
	fmt.Println(url)
	return nil
}

func init() {
	flag.StringVar(&keyword,
		"keyword",
		"",
		"The keyword which will be used to search")
	flag.IntVar(&limit,
		"limit",
		200,
		"The number of image will be transfered")
	flag.IntVar(&thread,
		"thread",
		10,
		"The number of thread which is used to do pulling/pushing")
}

func main() {
	flag.Parse()

	downloader := gg.NewDownloadImages(
		[]int{thread, thread},
		[]gg.PushHandler{pushImageToAssets},
		[]gg.PullHandler{pullImageToAssets},
	)
	if len(keyword) > 0 {
		downloader.Bulk(keyword, limit)
	}
}
