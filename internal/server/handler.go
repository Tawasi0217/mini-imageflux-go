package server

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"mini-imageflux-go/internal/imageproc"
)

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

	resp, err := http.Get(params.URL)
	if err != nil {
		log.Println("failed to fetch origin image:", err)
		http.Error(w, "failed to fetch origin image", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("origin returned status: %d", resp.StatusCode), http.StatusBadGateway)
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/gif" {
		http.Error(w, "unsupported origin content-type: "+contentType, http.StatusBadRequest)
		return
	}

	img, inputFormat, err := imageproc.Decode(resp.Body)
	if err != nil {
		log.Println("failed to decode image:", err)
		http.Error(w, "failed to decode image", http.StatusBadRequest)
		return
	}

	resized := imageproc.ResizeByWidth(img, params.Width)

	var buf bytes.Buffer

	if err := imageproc.EncodeJPEG(&buf, resized, 85); err != nil {
		log.Println("failed to encode jpeg:", err)
		http.Error(w, "failed to encode jpeg", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.Header().Set("X-Image-Proxy-Input-Format", inputFormat)
	w.Header().Set("X-Image-Proxy-Output-Format", "jpeg")
	w.Header().Set("X-Image-Proxy-Width", strconv.Itoa(params.Width))

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