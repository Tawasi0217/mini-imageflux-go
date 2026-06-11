package imageproc

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
)

func Decode(r io.Reader) (image.Image, string, error) {
	return image.Decode(r)
}