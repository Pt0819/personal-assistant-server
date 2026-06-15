package storage

import (
	"context"
	"fmt"
	"io"

	"personal-assistant-server/config"
)

type FileStorage interface {
	Upload(ctx context.Context, key string, reader io.Reader, contentType string) (string, error)
}

func New(cfg config.Oss) (FileStorage, error) {
	switch cfg.Type {
	case "aliyun":
		return newAliyunOSS(cfg)
	case "local":
		return newLocal(cfg)
	default:
		return nil, fmt.Errorf("unsupported oss type: %s", cfg.Type)
	}
}
