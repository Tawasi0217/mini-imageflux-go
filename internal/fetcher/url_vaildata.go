package fetcher

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

func ValidateOriginURL(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("url scheme must be http or https")
	}

	if parsedURL.Hostname() == "" {
		return fmt.Errorf("url host is required")
	}

	host := strings.ToLower(parsedURL.Hostname())

	if host == "localhost" {
		return fmt.Errorf("localhost is not allowed")
	}

	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() || ip.IsUnspecified() {
			return fmt.Errorf("local ip address is not allowed")
		}
	}

	return nil
}