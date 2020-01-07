package service

import (
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

type ErrorableBuiltServices struct {
	Services map[string]BuiltService
	Error    error
}

type Listener struct {
	reader    *Reader
	watcher   *fsnotify.Watcher
	Channel   <-chan ErrorableBuiltServices
	writeChan chan ErrorableBuiltServices
	closeChan chan struct{}
	syncChan  chan struct{}
}

func (l *Listener) readBuiltServices() ErrorableBuiltServices {
	services, err := l.reader.Read()
	return ErrorableBuiltServices{services, err}
}

func (l *Listener) AddCompose(path string) {
	l.reader.AddCompose(path)
	l.syncChan <- struct{}{}
	l.watcher.Add(path)
}

func (l *Listener) Close() error {
	l.closeChan <- struct{}{}
	return l.watcher.Close()
}

func (l *Listener) run() {
loop:
	for {
		select {
		case _, ok := <-l.watcher.Events:
			if !ok {
				break loop
			}
			l.writeChan <- l.readBuiltServices()
		case <-l.syncChan:
			l.writeChan <- l.readBuiltServices()
		case err, ok := <-l.watcher.Errors:
			if !ok {
				break loop
			}
			l.writeChan <- ErrorableBuiltServices{nil, err}
		case <-l.closeChan:
			break loop
		}
	}
	close(l.writeChan)
}

func NewServiceProvider() (*Listener, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create file watcher")
	}
	wc := make(chan ErrorableBuiltServices)
	l := &Listener{
		watcher:   w,
		Channel:   wc,
		writeChan: wc,
		closeChan: make(chan struct{}),
		syncChan:  make(chan struct{}),
	}
	go l.run()
	return l, nil
}
