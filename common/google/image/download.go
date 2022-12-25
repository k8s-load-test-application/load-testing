package image

import (
	"errors"
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
	// @TODO: please implement search image from google function
	return nil, errors.New("")
}

func (self *downloadImagesImpl) Bulk(keyword string) (int, error) {
	// @TODO: please implement bulk download function
	return 0, errors.New("Not implemented")
}

func New() DownloadImages {
	return &downloadImagesImpl{}
}
