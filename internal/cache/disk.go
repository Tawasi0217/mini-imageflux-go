package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

const CacheDir = ".cache/images"

type KeyParams struct {
	URL     string
	Width   int
	Format  string
	Quality int
}

func GenerateKey(params KeyParams) string {
	raw := fmt.Sprintf(
		"url=%s&w=%d&format=%s&q=%d",
		params.URL,
		params.Width,
		params.Format,
		params.Quality,
	)

	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

func Path(key string, format string) string {
	ext := format

	if ext == "jpeg" {
		ext = "jpg"
	}

	return filepath.Join(CacheDir, key+"."+ext)
}

func Exists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

func Read(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func Write(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}