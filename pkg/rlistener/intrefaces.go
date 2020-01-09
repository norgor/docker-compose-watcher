package rlistener

const (
	Create = Operation(1 << iota)
	Write
	Remove
	Rename
	Chmod
)

type Operation int

type WatcherMsg struct {
	Path      string
	Operation Operation
	Error     error
}

type Watcher interface {
	AddFolder(path string) error
	Channel() <-chan WatcherMsg
	Close() error
}
