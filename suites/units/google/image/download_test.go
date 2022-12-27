package main

import (
	"testing"

	gg "github.com/k8s-load-test-application/load-testing/common/google/image"
)

func TestSearch(t *testing.T) {
	dd := gg.NewDownloadImages([]int{10, 10}, nil, nil)
	l, err := dd.Search("jav")
	if len(l) == 0 || err != nil {
		t.Fatalf(`dd.Search() return empty list, %v`, err)
	}
}

func TestBulk(t *testing.T) {
	dd := gg.NewDownloadImages([]int{10, 10}, nil, nil)
	cnt, err := dd.Bulk("jav", 200)

	if cnt == 0 || err != nil {
		t.Fatalf(`dd.Bulk() return fetch fail, %v`, err)
	}
}
