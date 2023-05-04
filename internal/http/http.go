package internalhttp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

// file downloading methods using http libraries.

const acceptRangeHeader = "Accept-Ranges"

var (
	downloadError  DownloadError = `downloadError`
	cancelError    DownloadError = `cancelDownloadError`
	ErrCancelCopy                = errors.New(`cancelled by context`) // ErrCancelCopy Error occur by cancel
	copyBufferSize               = 32 * 1024
)

// DownloadError is a string used for context value key.
type DownloadError string

// getting url's head information, mostly for getting file size from Content-Length.
func getHead(url string) (*http.Response, error) {
	resp, err := http.Head(url)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// get content-length from header
func GetFileSizeAndResumable(url string) (int64, bool, error) {
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
func DownloadFile(ctx context.Context, url string, localFilePath string, downloadedBytes chan int, useResume bool, filesize int64, log func(param ...interface{})) {
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

// file download takes time if the file size was large.
// so instead of using io package copy, I made simple cancellable copy method.

func copyBuffer(ctx context.Context, dst io.Writer, src io.Reader, buf []byte) (written int64, err error) {
	if buf == nil { //default buffer size
		buf = make([]byte, copyBufferSize)
	}
loop:
	for {
		select {
		case <-ctx.Done():
			return written, ErrCancelCopy
		default:
			nr, er := src.Read(buf)
			if nr > 0 {
				nw, ew := dst.Write(buf[0:nr])
				if nw > 0 {
					written += int64(nw)
				}
				if ew != nil {
					err = ew
					break loop
				}
				if nr != nw {
					err = io.ErrShortWrite
					break loop
				}
			}
			if er != nil {
				if er != io.EOF {
					err = er
				}
				break loop
			}
		}
	}
	return written, err
}

// helper functions to resume file.

// check start point of resume file
func GetFileStartOffset(localfilePath string) (int64, error) {
	f, err := os.Stat(localfilePath)
	if err != nil {
		return 0, err
	}
	if f.IsDir() {
		return 0, errors.New(localfilePath + ` is directory. Not a file`)
	}
	return f.Size(), nil
}

// small files should not use resume
func IsFileShouldResume(contentLength int64) bool {
	return contentLength >= int64(copyBufferSize*1000)
}

// exmaple Range: bytes=0-1023
func rangeHeaderValue(file *os.File, currentLocalFileSize int64, contentLength int64) string {
	var begin int64
	if IsFileShouldResume(contentLength) {
		// while process may killed suddenly, last buffer of the file has possibility to be broken. so over write last buffer.
		modChunk := currentLocalFileSize % int64(copyBufferSize)
		begin = currentLocalFileSize - modChunk
		file.Seek(begin, 0)
	}
	return fmt.Sprintf(`bytes=%d-%d`, begin, contentLength)
}

// find download target file and its size to know the progress of download
func setupDownloadFile(localPath string, useResume bool) (*os.File, int64, error) {
	offset, err := GetFileStartOffset(localPath)
	var file *os.File
	if err != nil && os.IsNotExist(err) {
		file, err = os.Create(localPath)
		return file, 0, err
	}
	// use file that already exists
	file, err = os.OpenFile(localPath, os.O_RDWR, os.ModeExclusive)
	if err != nil {
		// can't open file for some reason
		return nil, 0, err
	}
	return file, offset, nil
}
