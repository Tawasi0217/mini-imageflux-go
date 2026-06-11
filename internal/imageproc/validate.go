package imageproc

import (
	"fmt"
	"image"
)

const MaxImagePixels = 25_000_000

func ValidateImageSize(img image.Image) error {
	bounds := img.Bounds()

	width := bounds.Dx()
	height := bounds.Dy()

	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid image dimensions: %dx%d", width, height)
	}

	pixels := width * height

	if pixels > MaxImagePixels {
		return fmt.Errorf("image dimensions are too large: %dx%d = %d pixels", width, height, pixels)
	}

	return nil
}