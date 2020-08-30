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
- Most Simple Usage. Download URL file to local File Example. 
```
	fdl := filedownloader.New(nil)
	user, _ := user.Current()
	err := fdl.SimpleFileDownload(`https://golang.org/pkg/net/http/`, user.HomeDir+`/fuso.html`)
	if err != nil {
		log.Println(err)
	}
```