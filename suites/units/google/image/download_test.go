package main

import (
	"testing"

	gg "github.com/k8s-load-test-application/load-testing/common/google/image"
)

func TestSearch(t *testing.T) {
	dd := gg.NewDownloadImages(10, 10)
	l, err := dd.Search("jav")

	if len(l) == 0 || err != nil {
		t.Fatalf(`dd.Search() return empty list, %v`, err)
	}
}
