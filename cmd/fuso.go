package filedownloader

import (
	"context"
	"errors"
	"fmt"
	logger "log"
	"strconv"
	"sync"
	"time"

	ihttp "github.com/sysgoblin/filedownloader/internal/http"
)

const (
	StateReady       state = `ready`       // StateReady is first state of instance
	StateDownloading state = `downloading` // StateDownloading is when the download started
	StateDone        state = `done`        // StateDone is when the download has finished or cancelled
)

var (
	ErrDownload = errors.New(`file download error`) // ErrDownload error component of downloader
)

type state string

// FileDownloader main structure
type FileDownloader struct {
	Conf                   *Config
	TotalFilesSize         int64
	ProgressChan           chan float64               // 0.0 to 1.0 float value indicates progress of downloading
	DownloadBytesPerSecond chan int64                 // downloaded bytes in last second
	Err                    error                      // error object
	Cancel                 func()                     // cancel downloading, if this method is called.
	LogFunc                func(param ...interface{}) // logging function
	State                  state                      // downloading state of filedownloader
}

// Config filedownloader config
type Config struct {
	MaxDownloadThreads     int                        // limit of parallel downloading threads. Default value is 3
	MaxRetry               int                        // retry count of file downloading, when download fails default is 0
	DownloadTimeoutMinutes int                        // download timeout minutes, default is 60
	RequiresDetailProgress bool                       // If true you can receive progress value from ProgressChan and downloadBytesPerSecond
	LogFunc                func(param ...interface{}) // logging function
}

// Download target url to download and local path to be downloaded
type Download struct {
	URL           string // downloading file URL
	LocalFilePath string // local file path which URL file will be downloaded
}

// New creates file downloader
func New(config *Config) *FileDownloader {
	if config == nil {
		config = &Config{MaxDownloadThreads: 3, MaxRetry: 0, DownloadTimeoutMinutes: 60, RequiresDetailProgress: false}
	}
	if config.MaxDownloadThreads == 0 {
		panic(`Check Configuration again. You can't download file if MaxDownloadThreads is 0`)
	}
	instance := &FileDownloader{Conf: config}
	// set default logger if not configured log function is not set.
	if config.LogFunc == nil {
		instance.LogFunc = fdlLog
	} else {
		// external log function
		instance.LogFunc = config.LogFunc
	}
	// create progress channels
	if instance.Conf.RequiresDetailProgress {
		progress := make(chan float64, 10)
		speed := make(chan int64, 10)
		instance.ProgressChan = progress
		instance.DownloadBytesPerSecond = speed
	}
	instance.State = StateReady
	return instance
}

// SimpleFileDownload simply download url file to localPath
func (m *FileDownloader) SimpleFileDownload(url, localFilePath string) error {
	if m.State != StateReady {
		panic(`filedownloader has already started or done`)
	}
	m.State = StateDownloading
	d := Download{URL: url, LocalFilePath: localFilePath}
	var list []*Download
	list = append(list, &d)
	// very simple single file download
	m.downloadFiles(list)
	return m.Err
}

// MultipleFileDownload downloads multiple files at parallel in configured download threads.
func (m *FileDownloader) MultipleFileDownload(downloads []*Download) error {
	if m.State != StateReady {
		panic(`filedownloader has already started or done`)
	}
	m.State = StateDownloading
	m.downloadFiles(downloads)
	return m.Err
}

func (m *FileDownloader) downloadFiles(downloads []*Download) {
	defer func() {
		m.State = StateDone
	}()
	downloadFilesCnt := len(downloads)
	m.LogFunc(`Download Files: ` + strconv.Itoa(downloadFilesCnt))
	// context for cancel and timeout
	ctx, timeoutFunc := context.WithTimeout(context.Background(), time.Minute*time.Duration(m.Conf.DownloadTimeoutMinutes))
	defer timeoutFunc()
	// if the url allows head access and returns Content-Length, we can calculate progress of downloading files.
	var resumableUrls = make(map[string]*resumeInfo)
	for _, d := range downloads {
		size, resumable, err := ihttp.GetFileSizeAndResumable(d.URL)
		if err != nil || size < 0 {
			panic(`Could not get whole size of the downloading file. No progress value is available`)
		}
		m.TotalFilesSize += size
		resumableUrls[d.URL] = &resumeInfo{isResumable: resumable, contentLength: size}
	}
	// count up downloaded bytes from download goroutines
	var downloadedBytes = make(chan int)
	defer close(downloadedBytes)
	// observe progress
	m.progressObserver(ctx, downloadedBytes)
	m.LogFunc(fmt.Sprintf("Total Download Bytes:: %d", m.TotalFilesSize))
	// Limit maximum download goroutines since network resource is not inifinite.
	dlCond := sync.NewCond(&sync.Mutex{})
	currentThreadCnt := 0
	var wg sync.WaitGroup
	// download context
	ctx2, timeoutFunc := context.WithTimeout(ctx, time.Minute*time.Duration(m.Conf.DownloadTimeoutMinutes))
	defer timeoutFunc()
	ctx3, cancelFunc := context.WithCancel(ctx2)
	defer cancelFunc()
	m.Cancel = cancelFunc
	// Downlaoding Files
	for i := 0; i < downloadFilesCnt; i++ {
		url := downloads[i].URL
		localPath := downloads[i].LocalFilePath
		resume, ok := resumableUrls[url]
		useResume := resume.isResumable && ok
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer dlCond.Signal()
			ihttp.DownloadFile(ctx3, url, localPath, downloadedBytes, useResume, resume.contentLength, m.LogFunc)
		}()
		currentThreadCnt++
		// stop for loop when reached to max threads.
		dlCond.L.Lock()
		if m.Conf.MaxDownloadThreads < currentThreadCnt {
			m.LogFunc(`Cond locked. download goroutine reached to max`)
			dlCond.Wait()
			m.LogFunc(`Cond released. goes to next file download if more.`)
			currentThreadCnt--
		}
		dlCond.L.Unlock()
	}
	m.LogFunc(`Wait group is waiting for download.`)
	// wait for all download ends.
	wg.Wait()
	// at last get the context error
	m.Err = ctx.Err()
	m.LogFunc(`All Download Task Done.`)
}

func (m *FileDownloader) progressObserver(ctx context.Context, downloadedBytes <-chan int) {
	var totaloDownloadedBytes int64
	m.LogFunc(`Total File Size from HTTP head Info::` + strconv.Itoa(int(m.TotalFilesSize)))
	// every second, print how many bytes downloaded.
	ticker := time.NewTicker(time.Second)

	// init the chans for progress and speed
	m.ProgressChan = make(chan float64)
	m.DownloadBytesPerSecond = make(chan int64)

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
				m.LogFunc(fmt.Sprintf(`downloaded %d bytes per second, downloaded %d / %d`, sub, totaloDownloadedBytes, m.TotalFilesSize))
				lastProgress = totaloDownloadedBytes
				if m.Conf.RequiresDetailProgress {
					m.DownloadBytesPerSecond <- sub
					// send progress value to channel. progress should be between 0.0 to 1.0.
					p := float64(totaloDownloadedBytes) / float64(m.TotalFilesSize)
					m.ProgressChan <- p
				}
			case t := <-downloadedBytes:
				// m.LogFunc(`Incomming bytes :` + strconv.Itoa(t))
				totaloDownloadedBytes += int64(t)
			case <-ctx.Done():
				m.LogFunc(`Progress Observer Done.`)
				break LOOP
			default:
				// noting to do
			}
		}
		m.LogFunc(`Filedownloader progress observer finished`)
	}()
}

func fdlLog(param ...interface{}) {
	logger.Println(param...)
}

type resumeInfo struct {
	isResumable   bool
	contentLength int64
}
