package storage

import (
	"bytes"
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
	"github.com/satmaelstorm/filup/internal/domain/port"
	"github.com/satmaelstorm/filup/internal/infrastructure/config"
	"io"
	"strings"
)

type MinioS3 struct {
	client    *minio.Client
	cfg       config.S3Config
	ctx       context.Context
	metaCache port.MetaCacheController
}

var storageClient *MinioS3

func ProvideMinioS3(cfg config.Configuration, cc port.ContextProvider, cache port.MetaCacheController) (*MinioS3, error) {
	if nil == storageClient {
		storageClient = new(MinioS3)
		storageClient.cfg = cfg.Storage.S3
		storageClient.ctx = cc.Ctx()
		c, err := minio.New(storageClient.cfg.Endpoint, &minio.Options{
			Creds: credentials.NewStaticV4(
				storageClient.cfg.Credentials.Key,
				storageClient.cfg.Credentials.Secret,
				storageClient.cfg.Credentials.Token,
			),
			Secure: storageClient.cfg.UseSSL,
		})
		if err != nil {
			return nil, err
		}
		storageClient.client = c
		err = storageClient.ensureBuckets()
		if err != nil {
			return nil, err
		}
		storageClient.metaCache = cache
	}
	return storageClient, nil
}

func (m *MinioS3) ensureBuckets() error {
	if err := m.ensureBucket(m.cfg.Buckets.Final); err != nil {
		return err
	}

	if err := m.ensureBucket(m.cfg.Buckets.Parts); err != nil {
		return err
	}

	if err := m.ensureBucket(m.cfg.Buckets.Meta); err != nil {
		return err
	}

	return nil
}

func (m *MinioS3) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(m.ctx, m.cfg.GetTimeout())
}

func (m *MinioS3) ensureBucket(bucketName string) error {
	ctx, cancel := m.getContextTimeout()
	defer cancel()
	b, err := m.client.BucketExists(ctx, bucketName)
	if err != nil {
		return errors.Wrap(err, "ensureBucket ("+bucketName+") ")
	}
	if !b {
		err := m.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
			Region:        m.cfg.Region,
			ObjectLocking: false,
		})
		if err != nil {
			return errors.Wrap(err, "ensureBucket ("+bucketName+") ")
		}
	}

	return nil
}

func (m *MinioS3) putFileByReader(contentType, bucketName, fileName string, filesize int64, content io.Reader) error {
	ctx, cancel := m.getContextTimeout()
	defer cancel()
	_, err := m.client.PutObject(
		ctx,
		bucketName,
		fileName,
		content,
		filesize,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return errors.Wrap(err, "MinioS3.putFile.PutObject")
	}
	return nil
}

func (m *MinioS3) putFile(contentType, bucketName, fileName string, content []byte) error {
	ctx, cancel := m.getContextTimeout()
	defer cancel()
	buf := bytes.NewBuffer(content)
	_, err := m.client.PutObject(
		ctx,
		bucketName,
		fileName,
		buf,
		int64(buf.Len()),
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return errors.Wrap(err, "MinioS3.putFile.PutObject")
	}
	return nil
}

func (m *MinioS3) getFile(bucketName, fileName string) ([]byte, error) {
	ctx, cancel := m.getContextTimeout()
	defer cancel()
	object, err := m.client.GetObject(ctx, bucketName, fileName, minio.GetObjectOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "MinioS3.getFile.GetObject")
	}
	stat, err := object.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "MinioS3.getFile.ObjectStat")
	}
	buf := make([]byte, stat.Size)
	if _, err := io.ReadFull(object, buf); err != nil { //nolint:govet
		return nil, errors.Wrap(err, "MinioS3.getFile.ReadFull")
	}
	return buf, nil
}

func (m *MinioS3) PutMetaFile(fileName string, content []byte) error {
	err := m.putFile("text/plain", m.cfg.Buckets.Meta, fileName, content)
	if err != nil {
		return errors.Wrap(err, "PutMetaFile")
	}
	m.metaCache.Add(fileName, content)
	return nil
}

func (m *MinioS3) GetMetaFile(fileName string) ([]byte, error) {
	content, ok := m.metaCache.Get(fileName)
	if ok && nil != content {
		return content, nil
	}
	content, err := m.getFile(m.cfg.Buckets.Meta, fileName)
	if err != nil {
		return nil, errors.Wrap(err, "GetMetaFile")
	}
	return content, nil
}

func (m *MinioS3) PutFilePart(fullPartName string, filesize int64, content io.Reader) error {
	err := m.putFileByReader(
		"application/octet-stream",
		m.cfg.Buckets.Parts,
		fullPartName,
		filesize,
		content,
	)
	if err != nil {
		return errors.Wrap(err, "MinioS3.PutFilePart")
	}
	return nil
}

func (m *MinioS3) GetLoadedFilePartsNames(fileName string) ([]string, error) {
	ctx, cancel := m.getContextTimeout()
	defer cancel()
	resultChan := m.client.ListObjects(ctx, m.cfg.Buckets.Parts, minio.ListObjectsOptions{
		Prefix:    fileName,
		Recursive: true,
	})
	var result []string //nolint:prealloc
	for obj := range resultChan {
		if obj.Err != nil {
			return nil, errors.Wrap(obj.Err, "MinioS3.GetLoadedFilePartsNames")
		}
		result = append(result, obj.Key)
	}
	return result, nil
}

func (m *MinioS3) ComposeFileParts(destFileName string, fullPartsName []string, tags map[string]string) (port.PartsComposerResult, error) {
	objects := make([]minio.CopySrcOptions, len(fullPartsName))
	for i, fn := range fullPartsName {
		objects[i] = minio.CopySrcOptions{Bucket: m.cfg.Buckets.Parts, Object: fn}
	}
	ctx, cancel := m.getContextTimeout()
	defer cancel()
	dest := minio.CopyDestOptions{
		Bucket:      m.cfg.Buckets.Final,
		Object:      destFileName,
		ReplaceTags: true,
		UserTags:    tags,
	}
	ui, err := m.client.ComposeObject(ctx, dest, objects...)
	if err != nil {
		return nil, errors.Wrap(err, "MinioS3.ComposeFileParts")
	}
	result := ComposeResult{
		bucket: ui.Bucket,
		name:   ui.Key,
		size:   ui.Size,
	}
	return result, nil
}

func (m *MinioS3) RemoveMeta(fileName string) error {
	ctx, cancel := m.getContextTimeout()
	defer cancel()
	err := m.client.RemoveObject(ctx, m.cfg.Buckets.Meta, fileName, minio.RemoveObjectOptions{ForceDelete: true, GovernanceBypass: true})
	if err != nil {
		return errors.Wrap(err, "MinioS3.RemoveMeta")
	}
	return nil
}

func (m *MinioS3) RemoveParts(chunkNames []string) error {
	ctx, cancel := m.getContextTimeout()
	defer cancel()
	var errs []string
	opts := minio.RemoveObjectOptions{ForceDelete: true, GovernanceBypass: true}
	for _, name := range chunkNames {
		err := m.client.RemoveObject(ctx, m.cfg.Buckets.Parts, name, opts)
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New("MinioS3.RemoveParts: [" + strings.Join(errs, ",") + "]")
	}
	return nil
}

type ComposeResult struct {
	bucket string
	name   string
	size   int64
}

func (c ComposeResult) GetBucket() string {
	return c.bucket
}

func (c ComposeResult) GetName() string {
	return c.name
}

func (c ComposeResult) GetSize() int64 {
	return c.size
}
