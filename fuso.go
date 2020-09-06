package filedownloader

import (
	"context"
	"errors"
	logger "log"
	"strconv"
	"sync"
	"time"
)

// FileDownloader main structure
type FileDownloader struct {
	conf           *Config
	TotalFilesSize int64
	ProgressChan   chan float64 // 0.0 to 1.0 float value indicates progress of downloading
	err            error        // error object
}

// Config filedownloader config
type Config struct {
	MaxDownloadThreads     int  // limit of parallel downloading threads. Default value is 3
	MaxRetry               int  // retry count of file downloading, when download fails default is 0
	DownloadTimeoutMinutes int  // download timeout minutes, default is 60
	RequiresProgress       bool // If true you can receive progress value from ProgressChan
}

// ErrDownload error component of downloader
var ErrDownload = errors.New(`File Download Error`)

// New creates file downloader
func New(config *Config) *FileDownloader {
	if config == nil {
		config = &Config{MaxDownloadThreads: 3, MaxRetry: 0, DownloadTimeoutMinutes: 60, RequiresProgress: false}
	}
	return &FileDownloader{conf: config}
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
	log(`Download Files: ` + strconv.Itoa(downloadFilesCnt))
	// context for cancel and timeout
	ctx, timeoutFunc := context.WithTimeout(context.Background(), time.Minute*time.Duration(m.conf.DownloadTimeoutMinutes))
	defer timeoutFunc()

	// if the url allows head access and returns Content-Length, we can calculate progress of downloading files.
	for _, url := range urlSlices {
		size, err := getFileSize(url)
		if err != nil || size < 0 {
			log(`Could not get whole size of the downloading file. No progress value is available`)
			m.TotalFilesSize = 0
			break
		}
		m.TotalFilesSize += size
	}
	// count up downloaded bytes from download goroutines
	var downloadedBytes = make(chan int)
	defer close(downloadedBytes)

	// observe progress
	m.progressCalculator(ctx, downloadedBytes)

	// Limit maximum download goroutines since network resource is not inifinite.
	dlThreads := sync.NewCond(&sync.Mutex{})
	currentThreadCnt := 0
	var wg sync.WaitGroup
	// Downlaoding Files
	for i := 0; i < len(urlSlices); i++ {
		url := urlSlices[i]
		localPath := localPaths[i]
		wg.Add(1)
		go downloadFile(ctx, dlThreads, &wg, url, localPath, downloadedBytes)

		currentThreadCnt++

		// stop for loop when reached to max threads.
		if m.conf.MaxDownloadThreads <= currentThreadCnt {
			dlThreads.L.Lock()
			log(`Cond lock executed. download goroutine reached to maximum`)
			dlThreads.Wait()
			currentThreadCnt--
			dlThreads.L.Unlock()
		}
	}

	// wait for all download ends.
	wg.Wait()
	// at last get the context error
	m.err = ctx.Err()
}

func (m *FileDownloader) progressCalculator(ctx context.Context, downloadedBytes <-chan int) {
	var totaloDownloadedBytes int
	log(`Total File Size on Internet::` + strconv.Itoa(int(m.TotalFilesSize)))
	// show progress
	progress := make(chan float64, 10)
	m.ProgressChan = progress
	go func() {
		defer close(progress)
	LOOP:
		for {
			select {
			case t := <-downloadedBytes:
				log(`Incomming bytes :` + strconv.Itoa(t))
				totaloDownloadedBytes += t
				log(totaloDownloadedBytes)
				// send progress value to channel. progress should be between 0.0 to 1.0.
				if t > 0 && m.conf.RequiresProgress {
					p := float64(totaloDownloadedBytes) / float64(m.TotalFilesSize)
					log(`send progress ` + strconv.FormatFloat(p, 'f', 2, 64))
					progress <- p
				}
			case <-ctx.Done():
				log(`Progress Calculator Done.`)
				break LOOP
			default:
				// noting to do
			}
		}
		log(`Finish Calculator goroutine`)
	}()
}

func log(param ...interface{}) {
	logger.Println(param...)
}
