package filedownloader

import (
	"fmt"
	"log"
	"os/user"
	_ "strconv"
	_ "sync"
	"testing"
	"time"
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

func TestMultipleFilesDownload(t *testing.T) {
	fdl := New(nil)
	user, _ := user.Current()
	var urlSlices []string
	urlSlices = append(urlSlices, `https://files.hareruyamtg.com/img/goods/L/M21/EN/0001.jpg`)
	urlSlices = append(urlSlices, `https://files.hareruyamtg.com/img/goods/L/ELD/EN/BRAWL0329.jpg`)
	var localPathSlices []string
	localPathSlices = append(localPathSlices, user.HomeDir+`/ugin.jpg`)
	localPathSlices = append(localPathSlices, user.HomeDir+`/korvold.jpg`)
	// Download Progress Observer
	/**var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
	LOOP:
		for {
			select {
			case bytes, ok := <-fdl.ProgressChan:
				fmt.Println(`progress comming.` + strconv.FormatFloat(bytes, 'f', 2, 64))
				if ok == false {
					break LOOP
				}
			default:
				t.Log(`No input`)
			}
		}
	}()*/

	err := fdl.MultipleFileDownload(urlSlices, localPathSlices)
	if err != nil {
		t.Error(err)
	}

	// wait for result
	/** wg.Wait() */
}

func TestFloatProgressCalc(t *testing.T) {
	v := float64(123 / float64(177476))
	fmt.Println(v)
}

func TestExternalLogFunction(t *testing.T) {
	conf := Config{logfunc: myLogger, MaxDownloadThreads: 1, DownloadTimeoutMinutes: 3}
	fileDownloader := New(&conf)
	// downloading to use home
	user, _ := user.Current()
	fileDownloader.SimpleFileDownload(`https://golang.org/pkg/net/http/`, user.HomeDir+`/fuso.html`)
}

func TestCancelWhileDownloading(t *testing.T) {
	conf := Config{logfunc: myLogger, MaxDownloadThreads: 1, DownloadTimeoutMinutes: 3, MaxRetry: 3}
	fileDownloader := New(&conf)
	// downloading to use home
	user, _ := user.Current()
	go func() {
		// stops downloading after 100 seconds
		time.Sleep(100 * time.Second)
		// wait and cancel
		fileDownloader.Cancel()
	}()
	// test download file 512MB
	err := fileDownloader.SimpleFileDownload(`http://ipv4.download.thinkbroadband.com/512MB.zip`, user.HomeDir+`/512.zip`)
	if err != nil {
		t.Error(err)
	}
	if fileDownloader.err != nil {
		t.Error(fileDownloader.err)
	}
	t.Log(`Test Done`)
}

func TestMultiFileDownloadCancelWhileDownloading(t *testing.T) {
	conf := Config{logfunc: myLogger, MaxDownloadThreads: 1, DownloadTimeoutMinutes: 3, MaxRetry: 3}
	fileDownloader := New(&conf)
	// downloading to use home
	user, _ := user.Current()
	go func() {
		// stops downloading after 100 seconds
		time.Sleep(10 * time.Second)
		// wait and cancel
		fileDownloader.Cancel()
	}()
	var urlSlices = []string{"http://ipv4.download.thinkbroadband.com/512MB.zip", "http://ipv4.download.thinkbroadband.com/200MB.zip"}
	var fileSlices = []string{user.HomeDir + `/512.zip`, user.HomeDir + `/200.zip`}
	// test download file 512MB
	err := fileDownloader.MultipleFileDownload(urlSlices, fileSlices)
	if err != nil {
		t.Error(err)
	}
	if fileDownloader.err != nil {
		t.Error(fileDownloader.err)
	}
	t.Log(`Test Done`)
}

func myLogger(params ...interface{}) {
	log.Println(`debug ::`, params)
}
