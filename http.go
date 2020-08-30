package filedownloader

import (
	"context"
	"io"
	"net/http"
	"os"
	"sync"
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
func downloadFile(ctx context.Context, c *sync.Cond, url string, localFilePath string, downloadedBytes chan int) {
	defer c.Signal()
	file, err := os.Create(localFilePath)
	if err != nil {
		log(err)
		ctx = context.WithValue(ctx, downloadError, err)
		return
	}
	defer file.Close()

	resp, err := http.Get(url)
	readSource := &responseReader{Reader: resp.Body, readBytes: downloadedBytes}

	_, err = io.Copy(file, readSource)
	if err != nil {
		log(err)
		ctx = context.WithValue(ctx, downloadError, err)
	}
}

// DownloadError is a string used for context value key.
type DownloadError string

var downloadError DownloadError = `downloadError`

// responseReader http response reader with channels
type responseReader struct {
	io.Reader
	readBytes chan int // send read bytes to channel
}

func (m *responseReader) Read(p []byte) (int, error) {
	n, err := m.Reader.Read(p)
	m.readBytes <- n
	return n, err
}
