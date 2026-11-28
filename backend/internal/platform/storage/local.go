// Package storage хранит загруженные файлы на локальном диске и отдаёт их по /uploads/.
package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Local struct {
	dir     string
	baseURL string
}

func NewLocal(dir, baseURL string) (*Local, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("create uploads dir: %w", err)
	}
	return &Local{dir: dir, baseURL: strings.TrimRight(baseURL, "/")}, nil
}

// Save сохраняет файл под случайным именем в папке folder и возвращает его публичный URL.
func (s *Local) Save(_ context.Context, folder, originalName string, r io.Reader) (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate file name: %w", err)
	}
	name := hex.EncodeToString(raw) + strings.ToLower(filepath.Ext(originalName))

	dir := filepath.Join(s.dir, filepath.Clean("/"+folder))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("create folder: %w", err)
	}

	f, err := os.OpenFile(filepath.Join(dir, name), os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o640) //nolint:gosec // имя генерируется сервером
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	if _, err := io.Copy(f, r); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return "", fmt.Errorf("write file: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("close file: %w", err)
	}
	return s.baseURL + path.Join("/", folder, name), nil
}

// Delete удаляет файл по URL, который вернул Save. Чужие URL игнорируются.
func (s *Local) Delete(_ context.Context, url string) error {
	rel, ok := strings.CutPrefix(url, s.baseURL+"/")
	if !ok {
		return nil
	}
	p := filepath.Join(s.dir, filepath.Clean("/"+rel))
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

// Handler раздаёт файлы без листинга папок.
func (s *Local) Handler(prefix string) http.Handler {
	fs := http.FileServer(http.Dir(s.dir))
	return http.StripPrefix(prefix, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "" || strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		fs.ServeHTTP(w, r)
	}))
}
