package storage

import (
	"context"
	"fmt"
	"io"

	"personal-assistant-server/config"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type aliyunOSS struct {
	bucket    *oss.Bucket
	bucketURL string
	basePath  string
}

func newAliyunOSS(cfg config.Oss) (FileStorage, error) {
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("创建OSS客户端失败: %w", err)
	}

	bucket, err := client.Bucket(cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("获取OSS Bucket失败: %w", err)
	}

	return &aliyunOSS{
		bucket:    bucket,
		bucketURL: cfg.BucketURL,
		basePath:  cfg.BasePath,
	}, nil
}

func (a *aliyunOSS) Upload(ctx context.Context, key string, reader io.Reader, contentType string) (string, error) {
	fullKey := a.basePath + key
	options := []oss.Option{
		oss.ContentType(contentType),
		oss.ACL(oss.ACLPublicRead),
	}

	if err := a.bucket.PutObject(fullKey, reader, options...); err != nil {
		return "", fmt.Errorf("上传OSS失败: %w", err)
	}

	return fmt.Sprintf("%s/%s", a.bucketURL, fullKey), nil
}
