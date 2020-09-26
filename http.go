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
		return nil, err
	}
	return resp, nil
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
func downloadFile(ctx context.Context, c *sync.Cond, wg *sync.WaitGroup, url string, localFilePath string, downloadedBytes chan int, log func(param ...interface{})) {
	defer c.Signal()
	defer wg.Done()

	select {
	case <-ctx.Done():
		log(`Download Cancelled by context`)
		return
	default:
		file, err := os.Create(localFilePath)
		if err != nil {
			return
		}
		defer file.Close()
		// download file
		resp, err := http.Get(url)
		if err != nil {
			ctx = context.WithValue(ctx, downloadError, err)
			return
		}
		defer resp.Body.Close()
		readSource := &responseReader{Reader: resp.Body, readBytes: downloadedBytes}
		ctxcp := context.WithValue(ctx, file, nil)
		_, err = copyBuffer(ctxcp, file, readSource, nil)
		if err != nil {
			if err == ErrCancelCopy {
				ctx = context.WithValue(ctx, cancelError, err)
				log(`Download File Cancelled[` + url + `]`)
			} else {
				ctx = context.WithValue(ctx, downloadError, err)
			}
			return
		}
	}
	log(`Download File Done[` + url + `]`)
}

// DownloadError is a string used for context value key.
type DownloadError string

var downloadError DownloadError = `downloadError`

var cancelError DownloadError = `cancelDownloadError`

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
