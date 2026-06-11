package imageproc

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"image/jpeg"
	"io"

	"golang.org/x/image/draw"
)

func Decode(r io.Reader) (image.Image, string, error) {
	return image.Decode(r)
}

func ResizeByWidth(img image.Image, targetWidth int) image.Image {
	bounds := img.Bounds()

	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	if targetWidth <= 0 {
		return img
	}

	if targetWidth == originalWidth {
		return img
	}

	targetHeight := originalHeight * targetWidth / originalWidth

	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	draw.CatmullRom.Scale(
		dst,
		dst.Bounds(),
		img,
		bounds,
		draw.Over,
		nil,
	)

	return dst
}

func EncodeJPEG(w io.Writer, img image.Image, quality int) error {
	if quality <= 0 || quality > 100 {
		quality = 85
	}

	return jpeg.Encode(w, img, &jpeg.Options{
		Quality: quality,
	})
}