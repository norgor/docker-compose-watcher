package service

import (
	"io"
	"os"
)

var osOpen = func(name string) (io.Reader, error) {
	return os.Open(name)
}
