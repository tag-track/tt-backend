package objectstore

import (
	"Backend/internal/env"
	"Backend/internal/thumbnail"
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"log"
	"sync"
)

type MinioAdapter struct {
	client *minio.Client
	bucket string
}

func NewMinioAdapter() *MinioAdapter {
	envVars := env.GetStaticEnv()

	client, err := minio.New(
		fmt.Sprintf("%s:%v", envVars.MinioHost, envVars.MinioPort),
		&minio.Options{
			Creds:  credentials.NewStaticV2(envVars.MinioUser, envVars.MinioPassword, ""),
			Secure: false,
		},
	)

	if err != nil {
		log.Printf(fmt.Sprintf("%s:%v", envVars.MinioHost, envVars.MinioPort))
		log.Fatalln(err)
	}

	return &MinioAdapter{
		client: client,
		bucket: envVars.MinioBucket,
	}
}

func (m *MinioAdapter) UpsertBucket(ctx context.Context, name string) error {
	if exists, err := m.client.BucketExists(ctx, name); exists || err != nil {
		if err != nil {
			return err
		}
		return nil
	}
	if err := m.client.MakeBucket(ctx, name, minio.MakeBucketOptions{}); err != nil {
		return err
	}
	return nil
}

func (m *MinioAdapter) UploadImage(ctx context.Context, filename string, img []byte) error {

	if err := m.UpsertBucket(ctx, m.bucket); err != nil {
		return err
	}

	contentType := "content/jpeg"

	info, err := m.client.PutObject(
		ctx,
		m.bucket,
		filename,
		bytes.NewReader(img),
		int64(len(img)),
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)

	if err != nil {
		log.Printf("Unable to upload thumbnail: %v", err)
		return err
	}

	log.Printf("Uploaded image %s [size: %v]", filename, info.Size)
	return nil
}

func (m *MinioAdapter) UploadThumbnail(ctx context.Context, t *thumbnail.Thumbnails) error {

	nameMap := t.GetImageDataMap()

	wg := sync.WaitGroup{}
	wg.Add(len(*nameMap))

	var errFlag error

	for name, img := range *nameMap {
		go func(name string, img []byte) {
			if err := m.UploadImage(ctx, name, img); err != nil {
				errFlag = err
			}
			wg.Done()
		}(name, img)
	}

	wg.Wait()

	return errFlag
}

func (m *MinioAdapter) RetrieveImage(ctx context.Context, name string) (io.ReadCloser, error) {

	_, err := m.client.StatObject(ctx, m.bucket, name, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}

	obj, err := m.client.GetObject(ctx, m.bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return obj, nil
}
