package imageproc

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
)

func Encode(w io.Writer, img image.Image, format string, quality int) error {
	switch format {
	case "jpeg", "jpg":
		return EncodeJPEG(w, img, quality)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func EncodeJPEG(w io.Writer, img image.Image, quality int) error {
	if quality <= 0 || quality > 100 {
		quality = 85
	}

	return jpeg.Encode(w, img, &jpeg.Options{
		Quality: quality,
	})
}

func ContentType(format string) string {
	switch format {
	case "jpeg", "jpg":
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}