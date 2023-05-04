package filedownloader

import (
	"bufio"
	"log"
	_url "net/url"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

// helpers
func validateURL(url string) (string, error) {
	// validate the url is valid
	u, err := _url.Parse(url)
	if err != nil {
		return "", err
	}

	// get the final element of the path to use as the filename
	fn := filepath.Base(u.Path)
	if fn == "" {
		return "", err
	}
	return fn, nil
}

func readFile(filename string) ([]string, error) {
	// validate the file path provided exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Fatal(err)
	}

	urls := []string{}
	// read the file and add each line to a slice
	file, err := os.Open(filename)
	if err != nil {
		return urls, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return urls, err
	}
	return urls, nil
}

// basic wrapper for fuso to cli app to use
func GoDownload(ctx *cli.Context) error {
	url := ctx.String("url")
	file := ctx.String("file")
	tor := ctx.Bool("tor")
	threads := ctx.Int("threads")
	retries := ctx.Int("retries")
	timeout := ctx.Int("timeout")

	proxy := ""
	if tor {
		// create proxy string
		proxy = "socks5://127.0.0.1:9050"
	}

	config := &Config{
		MaxDownloadThreads:     threads,
		MaxRetry:               retries,
		DownloadTimeoutMinutes: timeout,
		RequiresDetailProgress: false,
		Proxy:                  proxy,
	}

	if url != "" {
		// validate the url is valid
		fn, err := validateURL(url)
		if err != nil {
			log.Fatal(err)
		}

		fdl := New(config)
		err = fdl.SimpleFileDownload(url, fn)
		if err != nil {
			log.Fatal(err)
		}
	} else if file != "" {
		// read the file and add each line to a slice
		urls, err := readFile(file)
		if err != nil {
			log.Fatal(err)
		}

		// download each url in the slice
		fdl := New(config)
		var downloadFiles []*Download
		for _, url := range urls {
			// validate the url is valid
			fn, err := validateURL(url)
			if err != nil {
				log.Fatal(err)
			}
			downloadFiles = append(downloadFiles, &Download{URL: url, LocalFilePath: fn})
		}
		err = fdl.MultipleFileDownload(downloadFiles)
		if err != nil {
			log.Fatal(err)
		}

	} else {
		log.Fatal("no url or file given")
	}
	return nil
}
