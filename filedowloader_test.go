package filedownloader

import (
	logger "log"
	"testing"
)

func TestFileHeader(t *testing.T) {
	val, err := getFileSize(`https://golang.org/doc/install?download=go1.15.windows-amd64.msi`)
	if err != nil {
		t.Error(err)
	}
	logger.Println(val)
}
