package server

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"strconv"

	"mini-imageflux-go/internal/fetcher"
	"mini-imageflux-go/internal/imageproc"
)

const maxImageWidth = 4096

type ImageHandler struct{}

type ImageParams struct {
	URL    string
	Width  int
	Format string
}

func NewImageHandler() *ImageHandler {
	return &ImageHandler{}
}

func (h *ImageHandler) HandleImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	params, err := parseImageParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := fetcher.FetchImage(params.URL)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	limitedBody := http.MaxBytesReader(w, resp.Body, fetcher.MaxImageBytes)

	img, inputFormat, err := imageproc.Decode(limitedBody)
	if err != nil {
		log.Println("failed to decode image:", err)
		http.Error(w, "failed to decode image or image is too large", http.StatusBadRequest)
		return
	}

	if err := imageproc.ValidateImageSize(img); err != nil {
		log.Println("invalid image size:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resized := imageproc.ResizeByWidth(img, params.Width)

	var buf bytes.Buffer

	if err := imageproc.Encode(&buf, resized, params.Format, 85); err != nil {
		log.Println("failed to encode image:", err)
		http.Error(w, "failed to encode image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-Image-Proxy-Input-Format", inputFormat)
	w.Header().Set("X-Image-Proxy-Input-Width", strconv.Itoa(img.Bounds().Dx()))
	w.Header().Set("X-Image-Proxy-Input-Height", strconv.Itoa(img.Bounds().Dy()))
	w.Header().Set("X-Image-Proxy-Output-Format", params.Format)
	w.Header().Set("X-Image-Proxy-Requested-Width", strconv.Itoa(params.Width))
	w.Header().Set("X-Image-Proxy-Output-Width", strconv.Itoa(resized.Bounds().Dx()))
	w.Header().Set("X-Image-Proxy-Output-Height", strconv.Itoa(resized.Bounds().Dy()))

	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Println("failed to write response:", err)
	}
}

func parseImageParams(r *http.Request) (ImageParams, error) {
	query := r.URL.Query()

	originURL := query.Get("url")
	widthText := query.Get("w")
	format := query.Get("format")

	if originURL == "" {
		return ImageParams{}, errors.New("url is required")
	}

	if widthText == "" {
		return ImageParams{}, errors.New("w is required")
	}

	width, err := strconv.Atoi(widthText)
	if err != nil {
		return ImageParams{}, errors.New("w must be number")
	}

	if width <= 0 {
		return ImageParams{}, errors.New("w must be greater than 0")
	}

	if width > maxImageWidth {
		return ImageParams{}, errors.New("w must be less than or equal to 4096")
	}

	if format == "" {
		format = "jpeg"
	}

	if format != "jpeg" {
		return ImageParams{}, errors.New("format must be jpeg for now")
	}

	return ImageParams{
		URL:    originURL,
		Width:  width,
		Format: format,
	}, nil
}