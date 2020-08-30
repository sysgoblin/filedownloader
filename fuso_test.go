package filedownloader

import (
	"testing"
	"time"
)

// filedownloader test

func TestSimpleSingleDownload(t *testing.T) {
	fdl := New(nil)
	err := fdl.SimpleFileDownload(`https://golang.org/doc/install?download=go1.15.windows-amd64.msi`, `D:\\fuso.msi`)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(30000)
}
