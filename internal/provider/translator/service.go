package translator

import (
	"docker-compose-watcher/pkg/dockercompose/service"
	"docker-compose-watcher/pkg/flatmapper"
	"docker-compose-watcher/pkg/provider"
)

const labelTag = "dcw"

// WatchedService is a service that is provided by the Provider.
type WatchedService struct {
	Name      string
	Directory string
	Path      string `dcw:"docker-compose-watcher.path"`
}

func translate(src map[string]service.LabelledService) map[string]WatchedService {
	m := make(map[string]WatchedService, len(src))
	for k, v := range src {
		v := flatmapper.MapToStruct(labelTag, v.Labels, &WatchedService{
			Name:      v.Name,
			Directory: v.Directory,
		})
		s := *v.(*WatchedService)
		m[k] = s
	}
	return m
}

// NewServiceTranslatorChannel creates a new channel that translates LabelledService
// maps to WatchedService maps.
func NewServiceTranslatorChannel(src <-chan provider.ReaderValueWithError) <-chan provider.ReaderValueWithError {
	dst := make(chan provider.ReaderValueWithError)
	go func() {
		for {
			v, ok := <-src
			if !ok {
				break
			}
			if v.Error != nil {
				dst <- provider.ReaderValueWithError{
					Error: v.Error,
				}
				continue
			}
			dst <- provider.ReaderValueWithError{
				Value: translate(v.Value.(map[string]service.LabelledService)),
			}
		}
		close(dst)
	}()
	return dst
}
