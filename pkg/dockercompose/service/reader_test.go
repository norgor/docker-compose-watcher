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
		want     map[string]LabelledService
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
							Labels: map[string]string{
								"name.subkey1": "foo1",
								"name.subkey2": "foo2",
							},
						},
						"#2": {
							Labels: map[string]string{
								"name.subkey": "bar",
							},
						},
					},
				}),
				"/mnt/x/bar.yaml": composeToYamlReader(compose{
					Version: "2.6",
					Services: map[string]composeService{
						"#3": {
							Labels: map[string]string{
								"name.subkey1": "foo1",
								"name.subkey2": "foo2",
							},
						},
						"#4": {
							Labels: map[string]string{
								"name.subkey": "bar",
							},
						},
					},
				}),
			}),
			want: map[string]LabelledService{
				"#1": LabelledService{
					Name: "#1",
					Labels: map[string]string{
						"name.subkey1": "foo1",
						"name.subkey2": "foo2",
					},
				},
				"#2": LabelledService{
					Name: "#2",
					Labels: map[string]string{
						"name.subkey": "bar",
					},
				},
				"#3": LabelledService{
					Name: "#3",
					Labels: map[string]string{
						"name.subkey1": "foo1",
						"name.subkey2": "foo2",
					},
				},
				"#4": LabelledService{
					Name: "#4",
					Labels: map[string]string{
						"name.subkey": "bar",
					},
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
							Labels: map[string]string{
								"name.subkey1": "foo1",
								"name.subkey2": "foo2",
							},
						},
						"#2": {
							Labels: map[string]string{
								"name.subkey": "bar",
							},
						},
					},
				}),
				"bar.yaml": composeToYamlReader(compose{
					Version: "3.0",
					Services: map[string]composeService{
						"#1": {
							Labels: map[string]string{
								"name.s1": "f1",
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
			want: map[string]LabelledService{},
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
			got, err := r.ReadLabels()
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
