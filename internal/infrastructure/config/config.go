package config

import (
	"github.com/google/uuid"
	"github.com/satmaelstorm/filup/internal/infrastructure/logs/logsEngine"
	"net/url"
	"time"
)

const (
	minChunkLength         = 1024 * 1024 * 5 //@see github.com/minio/minio-go/v7@v7.0.18/constants.go:24
	maxPartsCount          = 10000
	maxSinglePutObjectSize = 1024 * 1024 * 1024 * 5
	minPartSize            = 1024 * 1024 * 16
)

type Configuration struct {
	Http     HTTP
	Storage  Storage
	Queue    QueueEngine
	Logs     logsEngine.LogConfigs
	Uploader Uploader
}

func (c *Configuration) AfterLoad() {
	for idx, cfg := range c.Logs {
		if cfg.MetricsOpts.Name != "" {
			cfg.MetricsOpts.Namespace = ProjectMetricsNamespace
			cfg.MetricsOpts.Subsystem = "logs"
			c.Logs[idx] = cfg
		}
	}
	c.Uploader = c.Uploader.AfterLoad()
}

type HTTP struct {
	Port    string
	Timeout int
}

func (h *HTTP) GetTimeout() time.Duration {
	return time.Duration(h.Timeout) * time.Second
}

type StorageCredentials struct {
	Key    string
	Secret string
	Token  string
}

type Storage struct {
	Type string
	S3   S3Config
}

type S3Config struct {
	UseSSL      bool
	MaxLifeTime int
	Credentials StorageCredentials
	Buckets     StorageBuckets
	Endpoint    string
	Region      string
}

func (s *S3Config) GetTimeout() time.Duration {
	return time.Duration(s.MaxLifeTime) * time.Second
}

type StorageBuckets struct {
	Parts string
	Final string
	Meta  string
}

type QueueEngine struct {
	Type        string
	Uri         string
	MaxLifeTime int
}

func (q *QueueEngine) GetTimeout() time.Duration {
	return time.Duration(q.MaxLifeTime) * time.Second
}

type Uploader struct {
	InfoFieldName  string
	ChunkLength    int64
	UuidNodeId     string
	CallbackBefore string
	CallbackAfter  string
	HttpTimeout    int64

	parsedCallbackBefore *url.URL
	parsedCallbackAfter  *url.URL
	httpTimeout          time.Duration
}

func (u Uploader) GetHttpTimeout() time.Duration {
	return u.httpTimeout
}

func (u Uploader) GetCallbackBefore() *url.URL {
	return u.parsedCallbackBefore
}

func (u Uploader) GetCallbackAfter() *url.URL {
	return u.parsedCallbackAfter
}

func (u Uploader) GetChunkLength() int64 {
	return u.ChunkLength
}

func (u Uploader) GetInfoFieldName() string {
	return u.InfoFieldName
}

func (u Uploader) GetMaxPartsCount() int64 {
	return maxPartsCount
}

func (u Uploader) GetMaxPartSize() int64 {
	return maxSinglePutObjectSize
}

func (u Uploader) GetOptPartSize() int64 {
	return minPartSize
}

func (u Uploader) AfterLoad() Uploader {
	if minChunkLength > u.ChunkLength {
		u.ChunkLength = minChunkLength
	}

	if "" == u.InfoFieldName {
		u.InfoFieldName = "_uploader_info"
	}
	if "" != u.UuidNodeId {
		b := uuid.SetNodeID([]byte(u.UuidNodeId))
		if !b {
			panic("config value uploader.uuidNodeId must be more than 6 bytes")
		}
		uuid.SetClockSequence(-1)
	}

	u.httpTimeout = time.Duration(u.HttpTimeout) * time.Second

	var err error
	if u.CallbackBefore != "" {
		u.parsedCallbackBefore, err = url.Parse(u.CallbackBefore)
		if err != nil {
			panic("error in Uploader.AfterLoad() while parse CallbackBefore: " + err.Error())
		}
	}

	if u.CallbackAfter != "" {
		u.parsedCallbackAfter, err = url.Parse(u.CallbackAfter)
		if err != nil {
			panic("error in Uploader.AfterLoad() while parse CallbackAfter: " + err.Error())
		}
	}

	return u
}
