package fsnotify

import (
	"docker-compose-watcher/pkg/provider"
	"github.com/fsnotify/fsnotify"
)

// Watcher watcher for file changes.
type Watcher struct {
	*fsnotify.Watcher
	ch chan provider.WatcherMsg
}

// Chan returns the watcher channel.
func (w *Watcher) Chan() <-chan provider.WatcherMsg {
	return w.ch
}

// New creates a new fsnotify watcher.
func New() (provider.Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	c := make(chan provider.WatcherMsg)
	go func() {
	loop:
		for {
			select {
			case err, ok := <-w.Errors:
				if !ok {
					break loop
				}
				c <- provider.WatcherMsg{
					Path: "",
					Err:  err,
				}
			case e, ok := <-w.Events:
				if !ok {
					break loop
				}
				c <- provider.WatcherMsg{
					Path: e.Name,
					Err:  nil,
				}
			}
		}
		close(c)
	}()
	return &Watcher{
		w,
		c,
	}, nil
}
