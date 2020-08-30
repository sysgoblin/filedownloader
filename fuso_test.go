package filedownloader

import (
	"os/user"

	"testing"
)

// filedownloader test

func TestSimpleSingleDownload(t *testing.T) {
	fdl := New(nil)
	user, _ := user.Current()
	err := fdl.SimpleFileDownload(`https://golang.org/pkg/net/http/`, user.HomeDir+`/fuso.html`)
	if err != nil {
		t.Error(err)
	}
}
