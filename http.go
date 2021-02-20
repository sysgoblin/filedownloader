package filedownloader

import (
	"context"
	"io"
	"net/http"
)

// file downloading methods using http libraries.

const acceptRangeHeader = "Accept-Ranges"

// getting url's head information, mostly for getting file size from Content-Length.
func getHead(url string) (*http.Response, error) {
	resp, err := http.Head(url)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// get content-length from header
func getFileSizeAndResumable(url string) (int64, bool, error) {
	resp, err := getHead(url)
	if err != nil {
		return 0, false, err
	}
	var acceptResume bool
	if resp.Header.Get(acceptRangeHeader) == "" {
		acceptResume = false
	} else {
		acceptResume = true
	}
	return resp.ContentLength, acceptResume, nil
}

// Download Single File
func downloadFile(ctx context.Context, url string, localFilePath string, downloadedBytes chan int, useResume bool, filesize int64, log func(param ...interface{})) {
	select {
	case <-ctx.Done():
		log(`Download Cancelled by context`)
		return
	default:
		file, offset, err := setupDownloadFile(localFilePath, useResume)
		if err != nil {
			return
		}
		defer file.Close()
		r, err := http.NewRequestWithContext(ctx, `GET`, url, nil)
		if useResume {
			r.Header.Add(`Range`, rangeHeaderValue(file, offset, filesize))
			log(`Resume enabled, added download header::`, r.Header)
		}
		if err != nil {
			return
		}
		// download file
		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			ctx = context.WithValue(ctx, downloadError, err)
			return
		}
		defer resp.Body.Close()
		readSource := &responseReader{Reader: resp.Body, readBytes: downloadedBytes}
		_, err = copyBuffer(ctx, file, readSource, nil)
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
