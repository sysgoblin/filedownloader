package filedownloader

import (
	"context"
	"errors"
	"fmt"
	logger "log"
	"strconv"
	"sync"
	"time"
)

// FileDownloader main structure
type FileDownloader struct {
	conf                   *Config
	TotalFilesSize         int64
	ProgressChan           chan float64               // 0.0 to 1.0 float value indicates progress of downloading
	DownloadBytesPerSecond chan int64                 // downloaded bytes in last second
	err                    error                      // error object
	Cancel                 func()                     // cancel downloading, if this method is called.
	logfunc                func(param ...interface{}) // logging function
}

// Config filedownloader config
type Config struct {
	MaxDownloadThreads     int                        // limit of parallel downloading threads. Default value is 3
	MaxRetry               int                        // retry count of file downloading, when download fails default is 0
	DownloadTimeoutMinutes int                        // download timeout minutes, default is 60
	RequiresDetailProgress bool                       // If true you can receive progress value from ProgressChan and downloadBytesPerSecond
	logfunc                func(param ...interface{}) // logging function
}

// ErrDownload error component of downloader
var ErrDownload = errors.New(`File Download Error`)

// New creates file downloader
func New(config *Config) *FileDownloader {
	if config == nil {
		config = &Config{MaxDownloadThreads: 3, MaxRetry: 0, DownloadTimeoutMinutes: 60, RequiresDetailProgress: false}
	}
	if config.MaxDownloadThreads == 0 {
		panic(`Check Configuration again. You can't download file if MaxDownloadThreads is 0`)
	}
	instance := &FileDownloader{conf: config}
	// set default logger if not configured log function is not set.
	if config.logfunc == nil {
		instance.logfunc = fdlLog
	} else {
		// external log function
		instance.logfunc = config.logfunc
	}
	// create progress channels
	if instance.conf.RequiresDetailProgress {
		progress := make(chan float64, 10)
		speed := make(chan int64, 10)
		instance.ProgressChan = progress
		instance.DownloadBytesPerSecond = speed
	}
	return instance
}

// SimpleFileDownload simply download url file to localPath
func (m *FileDownloader) SimpleFileDownload(url, localFilePath string) error {
	var urlSlice []string
	urlSlice = append(urlSlice, url)
	var localPaths []string
	localPaths = append(localPaths, localFilePath)
	// very simple single file download
	m.downloadFiles(urlSlice, localPaths)
	return m.err
}

// MultipleFileDownload downloads multiple files at parallel in configured download threads.
func (m *FileDownloader) MultipleFileDownload(urls, localFilePaths []string) error {
	m.downloadFiles(urls, localFilePaths)
	return m.err
}

func (m *FileDownloader) downloadFiles(urlSlices []string, localPaths []string) {
	if len(urlSlices) != len(localPaths) {
		m.err = errors.New(`url count and local download file path must match`)
		return
	}

	downloadFilesCnt := len(urlSlices)
	m.logfunc(`Download Files: ` + strconv.Itoa(downloadFilesCnt))
	// context for cancel and timeout
	ctx, timeoutFunc := context.WithTimeout(context.Background(), time.Minute*time.Duration(m.conf.DownloadTimeoutMinutes))
	defer timeoutFunc()
	// if the url allows head access and returns Content-Length, we can calculate progress of downloading files.
	for _, url := range urlSlices {
		size, err := getFileSize(url)
		if err != nil || size < 0 {
			m.logfunc(`Could not get whole size of the downloading file. No progress value is available`)
			m.TotalFilesSize = 0
			break
		}
		m.TotalFilesSize += size
	}
	// count up downloaded bytes from download goroutines
	var downloadedBytes = make(chan int)
	defer close(downloadedBytes)
	m.logfunc(`Progress Calculator Started`)
	// observe progress
	m.progressObserver(ctx, downloadedBytes)
	m.logfunc(fmt.Sprintf("Total Download Bytes:: %d", m.TotalFilesSize))
	// Limit maximum download goroutines since network resource is not inifinite.
	dlCond := sync.NewCond(&sync.Mutex{})
	currentThreadCnt := 0
	var wg sync.WaitGroup
	// download context
	ctx2, timeoutFunc := context.WithTimeout(ctx, time.Minute*time.Duration(m.conf.DownloadTimeoutMinutes))
	defer timeoutFunc()
	ctx3, cancelFunc := context.WithCancel(ctx2)
	defer cancelFunc()
	m.Cancel = cancelFunc
	// Downlaoding Files
	for i := 0; i < len(urlSlices); i++ {
		url := urlSlices[i]
		localPath := localPaths[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer dlCond.Signal()
			downloadFile(ctx3, url, localPath, downloadedBytes, m.logfunc)
		}()
		currentThreadCnt++
		// stop for loop when reached to max threads.
		dlCond.L.Lock()
		if m.conf.MaxDownloadThreads < currentThreadCnt {
			m.logfunc(`Cond locked. download goroutine reached to max`)
			dlCond.Wait()
			m.logfunc(`Cond released. goes to next file download if more.`)
			currentThreadCnt--
		}
		dlCond.L.Unlock()
	}
	m.logfunc(`Wait group is waiting for download.`)
	// wait for all download ends.
	wg.Wait()
	// at last get the context error
	m.err = ctx.Err()
	m.logfunc(`All Download Task Done.`)
}

func (m *FileDownloader) progressObserver(ctx context.Context, downloadedBytes <-chan int) {
	var totaloDownloadedBytes int64
	m.logfunc(`Total File Size from HTTP head Info::` + strconv.Itoa(int(m.TotalFilesSize)))
	// every second, print how many bytes downloaded.
	ticker := time.NewTicker(time.Second)
	go func() {
		defer close(m.ProgressChan)
		defer close(m.DownloadBytesPerSecond)
		defer ticker.Stop()
		var lastProgress int64
	LOOP:
		for {
			select {
			case <-ticker.C:
				sub := totaloDownloadedBytes - lastProgress
				m.logfunc(fmt.Sprintf(`downloaded %d bytes per second, downloaded %d / %d`, sub, totaloDownloadedBytes, m.TotalFilesSize))
				lastProgress = totaloDownloadedBytes
				if m.conf.RequiresDetailProgress {
					m.DownloadBytesPerSecond <- sub
					// send progress value to channel. progress should be between 0.0 to 1.0.
					p := float64(totaloDownloadedBytes) / float64(m.TotalFilesSize)
					m.ProgressChan <- p
				}
			case t := <-downloadedBytes:
				// m.logfunc(`Incomming bytes :` + strconv.Itoa(t))
				totaloDownloadedBytes += int64(t)
			case <-ctx.Done():
				m.logfunc(`Progress Observer Done.`)
				break LOOP
			default:
				// noting to do
			}
		}
		m.logfunc(`Filedownloader progress observer finished`)
	}()
}

func fdlLog(param ...interface{}) {
	logger.Println(param...)
}
