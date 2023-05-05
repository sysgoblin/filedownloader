package main

import (
	"os"

	filedownloader "github.com/sysgoblin/filedownloader/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{}
	app.Name = "filedownloader"
	app.Usage = "download file(s) from provided url(s)"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "url",
			Usage: "url to download",
		},
		&cli.StringFlag{
			Name:  "file",
			Usage: "file containing urls to download",
		},
		&cli.BoolFlag{
			Name:  "tor",
			Value: false,
			Usage: "download the given url through tor",
		},
		&cli.IntFlag{
			Name:  "threads",
			Value: 3,
			Usage: "number of threads to use for downloading from multiple urls",
		},
		&cli.IntFlag{
			Name:  "retries",
			Value: 0,
			Usage: "number of retries to attempt when downloading",
		},
		&cli.IntFlag{
			Name:  "timeout",
			Value: 60,
			Usage: "number of minutes to download before timing out",
		},
	}
	app.Action = func(ctx *cli.Context) error {
		// if both url and file are given, return error stating they both cannot be used
		if ctx.String("url") != "" && ctx.String("file") != "" {
			return cli.Exit("cannot use both url and file flags", 1)
		}
		filedownloader.GoDownload(ctx)
		return nil
	}
	app.Run(os.Args)
}
