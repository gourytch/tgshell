// +build windows
package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/vova616/screenshot"
)

func scrot(lossless bool) (data []byte, err error) {
	imgdata, err := screenshot.CaptureScreen()

	if err != nil {
		return
	}

	img := image.Image(imgdata)
	var buf bytes.Buffer
	if lossless {
		err = png.Encode(&buf, img)
	} else {
		err = jpeg.Encode(&buf, img, nil)
	}
	if err != nil {
		return
	}
	data = buf.Bytes()
	return
}
