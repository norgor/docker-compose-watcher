package serviceprovider

import (
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

type ErrorableBuiltServices struct {
	Services map[string]BuiltService
	Error    error
}

type ServiceProvider struct {
	files     []string
	watcher   *fsnotify.Watcher
	Channel   <-chan ErrorableBuiltServices
	writeChan chan ErrorableBuiltServices
	closeChan chan struct{}
	syncChan  chan struct{}
}

func (sp *ServiceProvider) readBuiltServices() ErrorableBuiltServices {
	services := make(map[string]BuiltService)
	for _, file := range sp.files {
		file, err := os.Open(file)
		if err != nil {
			return ErrorableBuiltServices{nil, err}
		}
		s, err := readServices(file)
		if err != nil {
			return ErrorableBuiltServices{nil, err}
		}
		for _, v := range s {
			if _, ok := services[v.Name]; ok {
				return ErrorableBuiltServices{
					nil,
					errors.Errorf("service '%s' is defined multiple times"),
				}
			}
			services[v.Name] = v
		}
	}
	return ErrorableBuiltServices{services, nil}
}

func (sp *ServiceProvider) AddCompose(path string) {
	sp.files = append(sp.files, path)
	sp.syncChan <- struct{}{}
	sp.watcher.Add(path)
}

func (sp *ServiceProvider) Close() error {
	sp.closeChan <- struct{}{}
	return sp.watcher.Close()
}

func (sp *ServiceProvider) run() {
loop:
	for {
		select {
		case _, ok := <-sp.watcher.Events:
			if !ok {
				break loop
			}
			sp.writeChan <- sp.readBuiltServices()
		case <-sp.syncChan:
			sp.writeChan <- sp.readBuiltServices()
		case err, ok := <-sp.watcher.Errors:
			if !ok {
				break loop
			}
			sp.writeChan <- ErrorableBuiltServices{nil, err}
		case <-sp.closeChan:
			break loop
		}
	}
	close(sp.writeChan)
}

func NewServiceProvider() (*ServiceProvider, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create file watcher")
	}
	wc := make(chan ErrorableBuiltServices)
	sp := &ServiceProvider{
		watcher:   w,
		Channel:   wc,
		writeChan: wc,
		closeChan: make(chan struct{}),
		syncChan:  make(chan struct{}),
	}
	go sp.run()
	return sp, nil
}
