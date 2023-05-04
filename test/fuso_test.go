package test

import (
	"fmt"
	"log"
	"os/user"
	_ "strconv"
	_ "sync"
	"testing"
	"time"

	fd "github.com/sysgoblin/filedownloader/cmd"
	ihttp "github.com/sysgoblin/filedownloader/internal/http"
)

func TestSimpleSingleDownload(t *testing.T) {
	fdl := fd.New(nil)
	user, _ := user.Current()
	err := fdl.SimpleFileDownload(`https://files.hareruyamtg.com/img/goods/L/M21/EN/0001.jpg`, user.HomeDir+`/ugin.jpg`)
	if err != nil {
		t.Error(err)
	}
}

func TestMultipleFilesDownload(t *testing.T) {
	fdl := fd.New(nil)
	user, _ := user.Current()
	// Download Progress Observer
	var downloadFiles []*fd.Download
	downloadFiles = append(downloadFiles, &fd.Download{URL: `https://files.hareruyamtg.com/img/goods/L/M21/EN/0001.jpg`, LocalFilePath: user.HomeDir + `/ugin.jpg`})
	downloadFiles = append(downloadFiles, &fd.Download{URL: `https://files.hareruyamtg.com/img/goods/L/ELD/EN/BRAWL0329.jpg`, LocalFilePath: user.HomeDir + `/korvold.jpg`})
	err := fdl.MultipleFileDownload(downloadFiles)
	if err != nil {
		t.Error(err)
	}
}

func TestFloatProgressCalc(t *testing.T) {
	v := float64(123 / float64(177476))
	fmt.Println(v)
}

func TestExternalLogFunc(t *testing.T) {
	conf := fd.Config{LogFunc: myLogger, MaxDownloadThreads: 1, DownloadTimeoutMinutes: 3}
	fileDownloader := fd.New(&conf)
	// downloading to use home
	user, _ := user.Current()
	fileDownloader.SimpleFileDownload(`https://files.hareruyamtg.com/img/goods/L/M21/EN/0001.jpg`, user.HomeDir+`/ugin.jpg`)
}

func TestCancelWhileDownloading(t *testing.T) {
	conf := fd.Config{LogFunc: myLogger, MaxDownloadThreads: 1, DownloadTimeoutMinutes: 3, MaxRetry: 3}
	fileDownloader := fd.New(&conf)
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
	if fileDownloader.Err != nil {
		t.Error(fileDownloader.Err)
	}
	t.Log(`Test Done`)
}

func TestFileDownloadWithDetailedConfiguration(t *testing.T) {
	// default setting of RequiresDetailProgress is false, you need to set it true if you need download progress.
	conf := fd.Config{LogFunc: myLogger, MaxDownloadThreads: 1, DownloadTimeoutMinutes: 3, MaxRetry: 3, RequiresDetailProgress: true}
	fileDownloader := fd.New(&conf)

	done := make(chan int)
	// if you set RequiresDetailProgress = true, you can receive progress from channel
	go func() {
	LOOP:
		for {
			select {
			case speed := <-fileDownloader.DownloadBytesPerSecond:
				// DownloadBytesPerSecond Channel can receive how fast the download is running.
				log.Printf(`%d bytes/sec`, speed)
			case progress := <-fileDownloader.ProgressChan:
				// Progress Channel (ProgressChan) receives how much download has progressed.
				log.Printf(`%f percent has done`, progress*100) // ex. 10.5 percent has done
			case <-done:
				break LOOP // escape from forever loop
			}
		}
		log.Println(`end of Observe loop`)
	}()

	// downloading file to use home directory
	user, _ := user.Current()
	// test download file 512MB
	err := fileDownloader.SimpleFileDownload(`http://ipv4.download.thinkbroadband.com/512MB.zip`, user.HomeDir+`/512.zip`)
	if err != nil {
		t.Error(err)
		done <- 1
	}
	if fileDownloader.Err != nil {
		t.Error(fileDownloader.Err)
	}
	done <- 0
	t.Log(`Test Done`)
}

func TestMultiFileDownloadCancelWhileDownloading(t *testing.T) {
	conf := fd.Config{LogFunc: myLogger, MaxDownloadThreads: 1, DownloadTimeoutMinutes: 3, MaxRetry: 3}
	fileDownloader := fd.New(&conf)
	// downloading to use home
	user, _ := user.Current()
	go func() {
		// stops downloading after 100 seconds
		time.Sleep(10 * time.Second)
		// wait and cancel
		fileDownloader.Cancel()
	}()
	var downloadFiles []*fd.Download
	downloadFiles = append(downloadFiles, &fd.Download{URL: "http://ipv4.download.thinkbroadband.com/512MB.zip", LocalFilePath: user.HomeDir + `/512.zip`})
	downloadFiles = append(downloadFiles, &fd.Download{URL: "http://ipv4.download.thinkbroadband.com/200MB.zip", LocalFilePath: user.HomeDir + `/200.zip`})
	// test download file 512MB
	err := fileDownloader.MultipleFileDownload(downloadFiles)
	if err != nil {
		t.Error(err)
	}
	if fileDownloader.Err != nil {
		t.Error(fileDownloader.Err)
	}
	t.Log(`Test Done`)
}

func myLogger(params ...interface{}) {
	//log.Println(`debug ::`, params)
}

func TestFileExists(t *testing.T) {
	user, _ := user.Current()
	bytes, err := ihttp.GetFileStartOffset(user.HomeDir + `/512.zip`)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(bytes)
}
