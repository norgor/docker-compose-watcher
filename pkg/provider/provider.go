package provider

import (
	"fmt"
	"github.com/pkg/errors"
)

// ReaderValueWithError contains services or the error that
// occured when reading the services.
type ReaderValueWithError struct {
	Value ReaderValue
	Error error
}

// Provider listens to changes to Docker Compose files. When the files change,
// the files are re-read and the values are sent to the channel.
type Provider struct {
	reader  Reader
	watcher Watcher
	ch      chan ReaderValueWithError
	closeCh chan struct{}
	syncCh  chan struct{}
}

func (l *Provider) read() ReaderValueWithError {
	v, err := l.reader.Read()
	return ReaderValueWithError{v, err}
}

// Add adds a Docker Compose file to listen to changes on.
func (l *Provider) Add(path string) error {
	err := l.reader.Add(path)
	if err != nil {
		return err
	}
	select {
	case l.syncCh <- struct{}{}:
	default:
	}
	return l.watcher.Add(path)
}

// Sync triggers a synchronization (full re-read) of the provider.
func (l *Provider) Sync() {
	l.syncCh <- struct{}{}
}

// Close cleans up and closes the channels.
func (l *Provider) Close() error {
	l.closeCh <- struct{}{}
	err := l.reader.Close()
	if err != nil {
		return err
	}
	return l.watcher.Close()
}

// Channel returns the provider channel.
func (l *Provider) Channel() <-chan ReaderValueWithError {
	return l.ch
}

func (l *Provider) run() {
loop:
	for {
		select {
		case v, ok := <-l.watcher.Chan():
			if !ok {
				break loop
			}
			if v.Err != nil {
				l.ch <- ReaderValueWithError{nil, v.Err}
				continue loop
			}
			l.ch <- l.read()
		case <-l.syncCh:
			l.ch <- l.read()
		case <-l.closeCh:
			break loop
		}
	}
	close(l.ch)
}

// WatcherFactoryFunc is a factory function for creating watchers.
type WatcherFactoryFunc func() (Watcher, error)

// ReaderFactoryFunc is a factory function for creating readers.
type ReaderFactoryFunc func() (Reader, error)

// New creates a new Provider.
func New(readerFactory ReaderFactoryFunc, watcherFactory WatcherFactoryFunc) (*Provider, error) {
	if readerFactory == nil {
		return nil, fmt.Errorf("readerFactory cannot be nil")
	}
	if watcherFactory == nil {
		return nil, fmt.Errorf("watcherFactory cannot be nil")
	}

	w, err := watcherFactory()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create watcher")
	}
	r, err := readerFactory()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create reader")
	}
	wc := make(chan ReaderValueWithError)
	l := &Provider{
		reader:  r,
		watcher: w,
		ch:      wc,
		closeCh: make(chan struct{}),
		syncCh:  make(chan struct{}, 1),
	}
	go l.run()
	return l, nil
}
