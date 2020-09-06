# filedownloader
 Download File from Internet

# Multipurpose File Downloader for Go
- Easy to download files from internet
- Download with its progress
- Multiple Download with Multiple Thread

# FUSO
Library project code is FUSO.
File Util for Simple Object

![](resources/fuso.jpg)

## This project is sill work in progress.

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
