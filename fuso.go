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
	conf *Config
	ctx  context.Context
	mtx  sync.Mutex
}

// Config filedownloader config
type Config struct {
	MaxDownloadThreads     int // limit of parallel downloading threads. Default value is 3
	MaxRetry               int // retry count of file downloading, when download fails default is 0
	DownloadTimeoutMinutes int // download timeout minutes, default is 60
}

// ErrDownload error component of downloader
var ErrDownload = errors.New(`File Download Error`)

// New creates file downloader
func New(config *Config) *FileDownloader {
	if config == nil {
		config = &Config{MaxDownloadThreads: 3, MaxRetry: 0, DownloadTimeoutMinutes: 60}
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
	_, err := m.downloadFiles(urlSlice, localPaths)
	return err
}

/**
func (m *FileDownloader) SimpleFileDownloadWithProgress((url, localFilePath string) (<-chan int, error) {

}*/

func (m *FileDownloader) downloadFiles(urlSlices []string, localPaths []string) (<-chan float32, error) {
	if len(urlSlices) != len(localPaths) {
		return nil, errors.New(`url count and local download file path must match`)
	}
	// show progress
	progress := make(chan float32)

	downloadFilesCnt := len(urlSlices)
	log(`Download Files: ` + strconv.Itoa(downloadFilesCnt))
	// context for cancel and timeout
	ctx, timeoutFunc := context.WithTimeout(context.Background(), time.Minute*time.Duration(m.conf.DownloadTimeoutMinutes))
	defer timeoutFunc()
	m.ctx = ctx

	var totalFilesSize int64
	var fileProgressAvailable bool = true
	// if the url allows head access and returns Content-Length, we can calculate progress of downloading files.
	for _, url := range urlSlices {
		size, err := getFileSize(url)
		if err != nil || size < 0 {
			log(`Could not get whole size of the downloading file. No progress value is available`)
			fileProgressAvailable = false
			break
		}
		totalFilesSize += size
	}

	// Downlaoding Files
	for i := 0; i < len(urlSlices); i++ {
		url := urlSlices[i]
		localPath := localPaths[i]
		go downloadFile(m.ctx, url, localPath, progress)
	}

	// Progress is unavailable
	if !fileProgressAvailable {
		close(progress)
	}
	return progress, nil
}

func log(param ...interface{}) {
	logger.Println(param...)
}
