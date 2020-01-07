package service

import (
	"io"
	"os"
	"reflect"
	"testing"
)

func stubOpen(m map[string]struct {
	name   string
	reader io.Reader
	err    error
}) func(name string) (io.Reader, error) {
	return func(name string) (io.Reader, error) {
		v, ok := m[name]
		if !ok {
			return nil, os.ErrNotExist
		}
		return v.reader, v.err
	}
}

func TestReader_Read(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		fileOpen func(name string) (io.Reader, error)
		want     map[string]BuiltService
		wantErr  bool
	}{
		{
			name: "reads specified files",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader()
			for _, v := range tt.files {
				r.AddCompose(v)
			}
			got, err := r.Read()
			if (err != nil) != tt.wantErr {
				t.Errorf("Reader.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reader.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}
