package filedownloader

import (
	"context"
	"net/http"
)

// file downloading methods using http libraries.

// getting url's head information, mostly for getting file size from Content-Length.
func getHead(url string) (*http.Response, error) {
	resp, err := http.Head(url)
	if err != nil {
		log(err)
	}
	return resp, err
}

// get content-length from header
func getFileSize(url string) (int64, error) {
	resp, err := getHead(url)
	if err != nil {
		return 0, err
	}
	return resp.ContentLength, nil
}

// Download Single File
func downloadFile(ctx context.Context, url string, localFilePath string, progress <-chan float32) (int, error) {

	return 0, nil
}
