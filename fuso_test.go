package filedownloader

import (
	"fmt"
	"os/user"
	_ "strconv"
	_ "sync"
	"testing"
)

// filedownloader test

func TestSimpleSingleDownload(t *testing.T) {
	fdl := New(nil)
	user, _ := user.Current()
	err := fdl.SimpleFileDownload(`https://golang.org/pkg/net/http/`, user.HomeDir+`/fuso.html`)
	if err != nil {
		t.Error(err)
	}
}

func TestMultipleFilesDownload(t *testing.T) {
	fdl := New(nil)
	user, _ := user.Current()
	var urlSlices []string
	urlSlices = append(urlSlices, `https://files.hareruyamtg.com/img/goods/L/M21/EN/0001.jpg`)
	urlSlices = append(urlSlices, `https://files.hareruyamtg.com/img/goods/L/ELD/EN/BRAWL0329.jpg`)
	var localPathSlices []string
	localPathSlices = append(localPathSlices, user.HomeDir+`/ugin.jpg`)
	localPathSlices = append(localPathSlices, user.HomeDir+`/korvold.jpg`)
	// Download Progress Observer
	/**var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
	LOOP:
		for {
			select {
			case bytes, ok := <-fdl.ProgressChan:
				fmt.Println(`progress comming.` + strconv.FormatFloat(bytes, 'f', 2, 64))
				if ok == false {
					break LOOP
				}
			default:
				t.Log(`No input`)
			}
		}
	}()*/

	err := fdl.MultipleFileDownload(urlSlices, localPathSlices)
	if err != nil {
		t.Error(err)
	}

	// wait for result
	/** wg.Wait() */
}

func TestFloatProgressCalc(t *testing.T) {
	v := float64(123 / float64(177476))
	fmt.Println(v)
}
