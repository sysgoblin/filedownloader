# filedownloader
 Download File from Internet

# Multipurpose File Downloader for Go
- Easy to download files from internet
- Progress Channel to receive downloading progress
- Multiple Download with Multiple Thread
- Able to configure Logging Function

# FUSO
Library project code is FUSO.

File Util for Simple Object

![](resources/fuso.jpg)

# Import
import github.com/chixm/filedownloader

# How to use 
## Example1 Most Simple Usage. Download URL file to local File Example. 
```
	fdl := filedownloader.New(nil)
	user, _ := user.Current()
	err := fdl.SimpleFileDownload(`https://golang.org/pkg/net/http/`, user.HomeDir+`/fuso.html`)
	if err != nil {
		log.Println(err)
	}
```

## Example2: Multiple Files Download From Internet
```
	fdl := filedownloader.New(nil)
	user, _ := user.Current()
	// make a URL slice of downloading files
	var urlSlices []string
	urlSlices = append(urlSlices, `https://files.hareruyamtg.com/img/goods/L/M21/EN/0001.jpg`)
	urlSlices = append(urlSlices, `https://files.hareruyamtg.com/img/goods/L/ELD/EN/BRAWL0329.jpg`)
	// make a slice of LocalFile download paths.
	var localPathSlices []string
	localPathSlices = append(localPathSlices, user.HomeDir+`/ugin.jpg`)
	localPathSlices = append(localPathSlices, user.HomeDir+`/korvold.jpg`)

	err := fdl.MultipleFileDownload(urlSlices, localPathSlices)
	if err != nil {
		t.Error(err)
	}

```

## Self defined Log functions
FUSO writes log by Golang "log" library by default.
You can configure Config and set your own logger.

```
conf := Config{logfunc: myLogger}
	fileDownloader := New(&conf)
	// downloading to use home
	user, _ := user.Current()
	fileDownloader.SimpleFileDownload(`https://golang.org/pkg/net/http/`, user.HomeDir+`/fuso.html`)

...

func myLogger(params ...interface{}) {
	log.Println(`log prefix`, params)
}

```

## Configure RequiresDetailProgress And Receive Progress
You can show downloading progress and downloading speed if you need.
Set RequiresDetailProgress: true in Config and write channel receives progress data.

See example below,
```
	// default setting of RequiresDetailProgress is false, you need to set it true if you need download progress.
	conf := Config{logfunc: myLogger, MaxDownloadThreads: 1, DownloadTimeoutMinutes: 3, MaxRetry: 3, RequiresDetailProgress: true}
	fileDownloader := New(&conf)

	done := make(chan int)
	// if you set RequiresDetailProgress = true, you can receive progress from channel
	go func() {
	LOOP:
		for {
			select {
			case speed := <-fileDownloader.DownloadBytesPerSecond:
				// DownloadBytesPerSecond Channel can receive how fast the download is running.
				log.Println(fmt.Sprintf(`%d bytes/sec`, speed))
			case progress := <-fileDownloader.ProgressChan:
				// Progress Channel (ProgressChan) receives how much download has progressed.
				log.Println(fmt.Sprintf(`%f percent has done`, progress)) // ex. 10.5 percent has done
			case <-done:
				break LOOP // escape from forever loop
			}
		}
	}()

	// downloading file to use home directory
	user, _ := user.Current()
	// test download file 512MB
	err := fileDownloader.SimpleFileDownload(`http://ipv4.download.thinkbroadband.com/512MB.zip`, user.HomeDir+`/512.zip`)
	if err != nil {
		done <- 1
	}
	if fileDownloader.err != nil {
		log.Println(fileDownloader.err)
	}
	done <- 0
```