package filedownloader

import (
	"errors"
	"fmt"
	"os"
)

// helper functions to resume file.

// check start point of resume file
func getFileStartOffset(localfilePath string) (int64, error) {
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
func isFileShouldResume(contentLength int64) bool {
	if contentLength < int64(copyBufferSize*1000) {
		return false
	}
	return true
}

// exmaple Range: bytes=0-1023
func rangeHeaderValue(file *os.File, currentLocalFileSize int64, contentLength int64) string {
	var begin int64
	if isFileShouldResume(contentLength) {
		// while process may killed suddenly, last buffer of the file has possibility to be broken. so over write last buffer.
		modChunk := currentLocalFileSize % int64(copyBufferSize)
		begin = currentLocalFileSize - modChunk
		file.Seek(begin, 0)
	}
	return fmt.Sprintf(`bytes=%d-%d`, begin, contentLength)
}

// find download target file and its size to know the progress of download
func setupDownloadFile(localPath string, useResume bool) (*os.File, int64, error) {
	offset, err := getFileStartOffset(localPath)
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
