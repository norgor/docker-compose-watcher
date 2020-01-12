package reader

import (
	"docker-compose-watcher/pkg/dockercompose/service"
	"docker-compose-watcher/pkg/provider"
)

type readerImpl struct {
	r *service.Reader
}

func (r *readerImpl) Add(path string) error {
	r.r.Add(path)
	return nil
}

func (r *readerImpl) Close() error {
	return nil
}

func (r *readerImpl) Read() (provider.ReaderValue, error) {
	return r.r.ReadLabels()
}

// NewServiceReader creates a new docker compose service reader.
func NewServiceReader() (provider.Reader, error) {
	return &readerImpl{service.NewReader()}, nil
}
