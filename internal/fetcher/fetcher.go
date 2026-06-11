package fetcher

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const MaxImageBytes = 10 * 1024 * 1024 // 10MB

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("too many redirects")
		}

		if err := ValidateOriginURL(req.URL.String()); err != nil {
			return fmt.Errorf("redirect URL is not allowed: %w", err)
		}

		return nil
	},
}

func FetchImage(originURL string) (*http.Response, error) {
	if err := ValidateOriginURL(originURL); err != nil {
		return nil, err
	}

	resp, err := httpClient.Get(originURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch origin image: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("origin returned status: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !isSupportedImageContentType(contentType) {
		resp.Body.Close()
		return nil, fmt.Errorf("unsupported origin content-type: %s", contentType)
	}

	if resp.ContentLength > MaxImageBytes {
		resp.Body.Close()
		return nil, fmt.Errorf("origin image is too large: %d bytes", resp.ContentLength)
	}

	return resp, nil
}

func isSupportedImageContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "image/jpeg") ||
		strings.HasPrefix(contentType, "image/png") ||
		strings.HasPrefix(contentType, "image/gif")
}