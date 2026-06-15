package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"personal-assistant-server/config"
)

type localStorage struct {
	basePath string
	saveDir  string
}

func newLocal(cfg config.Oss) (FileStorage, error) {
	dir := filepath.Join("uploads", cfg.BasePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建本地存储目录失败: %w", err)
	}
	return &localStorage{
		basePath: cfg.BasePath,
		saveDir:  dir,
	}, nil
}

func (l *localStorage) Upload(ctx context.Context, key string, reader io.Reader, contentType string) (string, error) {
	filename := filepath.Join(l.saveDir, filepath.Base(key))

	f, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	return "/" + filename, nil
}
