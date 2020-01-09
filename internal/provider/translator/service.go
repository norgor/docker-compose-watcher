package translator

import (
	"docker-compose-watcher/pkg/dockercompose/service"
	"docker-compose-watcher/pkg/flatmapper"
	"docker-compose-watcher/pkg/rprovider"
)

const labelTag = "dcw"

type WatchedService struct {
	Name string
	Path string `dcw:docker-compose-watcher.path`
}

func translate(src map[string]service.LabelledService) map[string]WatchedService {
	m := make(map[string]WatchedService, len(src))
	for k, v := range src {
		m[k] = *flatmapper.MapToStruct(labelTag, v.Labels, &WatchedService{}).(*WatchedService)
	}
	return m
}

// NewServiceTranslatorChannel creates a new channel that translates LabelledService
// maps to WatchedService maps.
func NewServiceTranslatorChannel(src <-chan rprovider.ReaderValueWithError) <-chan rprovider.ReaderValueWithError {
	dst := make(chan rprovider.ReaderValueWithError)
	go func() {
		for {
			v, ok := <-src
			if !ok {
				break
			}
			if v.Error != nil {
				dst <- rprovider.ReaderValueWithError{
					Value: nil,
					Error: v.Error,
				}
				continue
			}

		}
		close(dst)
	}()
	return dst
}
