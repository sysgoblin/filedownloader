# godownload
A small cli go app to download file(s) from the internet.
Provide either a URL to a file to download, or a plain text file containing a list of URLs to process.

The downloader will first gather the total size of all files to download, and track progress. Downloading is resumable, so if the command is stopped while downloading a large file, running the command again will continue the download as long as the partially complete file is still present locally.

```
NAME:
   godownload - download file(s) from provided url(s)

USAGE:
   godownload [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --url value      url to download
   --file value     file containing a list of newline separated urls to download
   --tor            download the given url through local tor proxy (127.0.0.1:9050) (default: false)
   --threads value  number of threads to use for downloading from multiple urls (default: 3)
   --retries value  number of retries to attempt when downloading (default: 0)
   --timeout value  number of minutes to download before timing out (default: 60)
   --help, -h       show help
   --version, -v    print the version
```

A repurposed fork of https://github.com/chixm/filedownloader