package rprovider

// ReaderValue is a generic value from a reader.
type ReaderValue interface{}

// Adder is an interface for adding paths.
type Adder interface {
	Add(path string) error
}

// Closer is an interface for closing.
type Closer interface {
	Close() error
}

// Reader reads a generic value.
type Reader interface {
	Adder
	Closer
	Read() (ReaderValue, error)
}

// WatcherMsg is an event sent by the Watcher.
type WatcherMsg struct {
	Path string
	Err  error
}

// Watcher watcher for file changes.
type Watcher interface {
	Adder
	Closer
	Chan() <-chan WatcherMsg
}
