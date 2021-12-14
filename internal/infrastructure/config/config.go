package config

import (
	"github.com/satmaelstorm/filup/internal/infrastructure/logs/logsEngine"
	"time"
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
	Endpoint    string
	UseSSL      bool
	Region      string
	MaxLifeTime int
	Credentials StorageCredentials
}

func (s *Storage) GetTimeout() time.Duration {
	return time.Duration(s.MaxLifeTime) * time.Second
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
	CopyHeaders   []string
	MaxFileLength int64
}

func (u Uploader) GetCopyHeaders() []string {
	return u.CopyHeaders
}

func (u Uploader) GetMaxFileLength() int64 {
	return u.MaxFileLength
}
