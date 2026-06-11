package imageproc

import (
	"image"

	"golang.org/x/image/draw"
)

func ResizeByWidth(img image.Image, targetWidth int) image.Image {
	bounds := img.Bounds()

	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	if targetWidth <= 0 {
		return img
	}

	if targetWidth >= originalWidth {
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