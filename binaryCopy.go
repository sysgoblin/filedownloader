package filedownloader

import (
	"context"
	"errors"
	"io"
)

// file download takes time if the file size was large.
// so instead of using io package copy, I made simple cancellable copy method.

// ErrCancelCopy Error occur by cancel
var ErrCancelCopy = errors.New(`Cancelled by context`)

var copyBufferSize = 32 * 1024

func copyBuffer(ctx context.Context, dst io.Writer, src io.Reader, buf []byte) (written int64, err error) {
	if buf == nil { //default buffer size
		buf = make([]byte, copyBufferSize)
	}
loop:
	for {
		select {
		case <-ctx.Done():
			return written, ErrCancelCopy
		default:
			nr, er := src.Read(buf)
			if nr > 0 {
				nw, ew := dst.Write(buf[0:nr])
				if nw > 0 {
					written += int64(nw)
				}
				if ew != nil {
					err = ew
					break loop
				}
				if nr != nw {
					err = io.ErrShortWrite
					break loop
				}
			}
			if er != nil {
				if er != io.EOF {
					err = er
				}
				break loop
			}
		}
	}
	return written, err
}
