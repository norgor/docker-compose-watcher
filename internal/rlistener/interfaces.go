package rlistener

// Operations
const (
	Create = Operation(1 << iota)
	Write
	Remove
	Rename
	Chmod
)

// Operation is the operation type that can be performed on a file.
type Operation int

// WatcherMsg is a message from a Watcher.
type WatcherMsg struct {
	Path string
	Op   Operation
	Err  error
}

// Watcher provides an interface for watching directories.
type Watcher interface {
	RemDir(path string) error
	AddDir(path string) error
	Channel() <-chan WatcherMsg
	Close() error
}
