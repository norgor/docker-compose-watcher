package fsnotify

import (
	"docker-compose-watcher/pkg/rprovider"
	"github.com/fsnotify/fsnotify"
)

// Watcher watcher for file changes.
type Watcher struct {
	*fsnotify.Watcher
	ch chan rprovider.WatcherMsg
}

// Chan returns the watcher channel.
func (w *Watcher) Chan() <-chan rprovider.WatcherMsg {
	return w.ch
}

// New creates a new fsnotify watcher.
func New() (rprovider.Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	c := make(chan rprovider.WatcherMsg)
	go func() {
	loop:
		for {
			select {
			case err, ok := <-w.Errors:
				if !ok {
					break loop
				}
				c <- rprovider.WatcherMsg{
					Path: "",
					Err:  err,
				}
			case e, ok := <-w.Events:
				if !ok {
					break loop
				}
				c <- rprovider.WatcherMsg{
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
