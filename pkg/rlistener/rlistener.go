package rlistener

type Listener struct {
	w Watcher
}

type WatcherFactoryFunc func() (Watcher, error)

func New(watcherFactory WatcherFactoryFunc) error {
	wat
}
