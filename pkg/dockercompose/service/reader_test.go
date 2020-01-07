package service

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

type yamlReader struct {
	reader io.Reader
	err    error
}

type openMap map[string]yamlReader

func stubOpen(m openMap) func(name string) (io.Reader, error) {
	return func(name string) (io.Reader, error) {
		v, ok := m[name]
		if !ok {
			return nil, os.ErrNotExist
		}
		return v.reader, v.err
	}
}

func composeToYamlReader(a compose) yamlReader {
	d, err := yaml.Marshal(&a)
	if err != nil {
		return yamlReader{nil, err}
	}
	return yamlReader{bytes.NewReader(d), nil}
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
			name:  "reads specified files",
			files: []string{"/mnt/x/foo.yaml", "/mnt/x/bar.yaml"},
			fileOpen: stubOpen(openMap{
				"/mnt/x/foo.yaml": composeToYamlReader(compose{
					Version: "3.0",
					Services: map[string]composeService{
						"#1": {
							Build: "./foo1",
						},
						"#2": {
							Build: map[string]interface{}{
								"context": "./foo2",
							},
						},
					},
				}),
				"/mnt/x/bar.yaml": composeToYamlReader(compose{
					Version: "2.6",
					Services: map[string]composeService{
						"#3": {
							Build: "./bar1",
						},
						"#4": {
							Build: map[string]interface{}{
								"context": "./bar2",
							},
						},
					},
				}),
			}),
			want: map[string]BuiltService{
				"#1": BuiltService{
					Name: "#1",
					Path: "./foo1",
				},
				"#2": BuiltService{
					Name: "#2",
					Path: "./foo2",
				},
				"#3": BuiltService{
					Name: "#3",
					Path: "./bar1",
				},
				"#4": BuiltService{
					Name: "#4",
					Path: "./bar2",
				},
			},
		},
		{
			name:  "service defined multiple times in separate files",
			files: []string{"/mnt/x/foo.yaml", "bar.yaml"},
			fileOpen: stubOpen(openMap{
				"/mnt/x/foo.yaml": composeToYamlReader(compose{
					Version: "3.0",
					Services: map[string]composeService{
						"#1": {
							Build: "./foo1",
						},
						"#2": {
							Build: map[string]interface{}{
								"context": "./foo2",
							},
						},
					},
				}),
				"bar.yaml": composeToYamlReader(compose{
					Version: "3.0",
					Services: map[string]composeService{
						"#1": {
							Build: "./foo1",
						},
						"#2": {
							Build: map[string]interface{}{
								"context": "./foo2",
							},
						},
					},
				}),
			}),
			want:    nil,
			wantErr: true,
		},
		{
			name:  "file without services",
			files: []string{"/mnt/x/foo.yaml"},
			fileOpen: stubOpen(openMap{
				"/mnt/x/foo.yaml": composeToYamlReader(compose{
					Version: "3.0",
				}),
			}),
			want: map[string]BuiltService{},
		},
		{
			name:  "unsupported version",
			files: []string{"/mnt/x/foo.yaml"},
			fileOpen: stubOpen(openMap{
				"/mnt/x/foo.yaml": composeToYamlReader(compose{
					Version: "4.0",
				}),
			}),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldOsOpen := osOpen
			osOpen = tt.fileOpen
			defer func() {
				osOpen = oldOsOpen
			}()
			r := NewReader()
			for _, v := range tt.files {
				r.Add(v)
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
