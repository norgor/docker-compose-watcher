package service

import (
	"io"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// LabelledService is a Docker Compose service that is labelled.
type LabelledService struct {
	Name      string
	Directory string
	Labels    map[string]string
}

type composeService struct {
	Labels map[string]string `yaml:"labels"`
}

type compose struct {
	Version  string `yaml:"version"`
	Services map[string]composeService
}

type version struct {
	Major int
	Minor int
}

// Reader reads from the Docker Compose yaml files
type Reader struct {
	files []string
}

func parseVersion(ver string) (version, error) {
	// hacky and amazing way to do it
	floatVer, err := strconv.ParseFloat(ver, 32)
	if err != nil {
		return version{}, errors.Wrapf(err, "unable to parse version")
	}
	major := int(floatVer)
	minor := int((floatVer - float64(major)) * 10)
	return version{
		Major: major,
		Minor: minor,
	}, nil
}

func validateVersion(ver string) error {
	version, err := parseVersion(ver)
	if err != nil {
		return err
	}
	if version.Major < 1 || version.Major > 3 {
		return errors.New("invalid version")
	}
	return nil
}

func getServiceLabels(service *composeService) map[string]string {
	if service.Labels == nil {
		return make(map[string]string, 0)
	}
	return service.Labels
}

func readFromCompose(reader io.Reader) ([]LabelledService, error) {
	compose := compose{}
	decoder := yaml.NewDecoder(reader)
	if err := decoder.Decode(&compose); err != nil {
		return nil, err
	}
	if err := validateVersion(compose.Version); err != nil {
		return nil, err
	}
	return transformServices(&compose)
}

func transformServices(compose *compose) ([]LabelledService, error) {
	var services []LabelledService
	for serviceName, service := range compose.Services {
		services = append(services, LabelledService{
			Name:   serviceName,
			Labels: getServiceLabels(&service),
		})
	}
	return services, nil
}

// ReadLabels reads all the services and their labels from the Docker Compose files.
func (r *Reader) ReadLabels() (map[string]LabelledService, error) {
	services := make(map[string]LabelledService)
	for _, file := range r.files {
		f, err := osOpen(file)
		if err != nil {
			return nil, err
		}
		s, err := readFromCompose(f)
		if err != nil {
			return nil, err
		}
		d := filepath.Dir(file)
		for _, v := range s {
			if _, ok := services[v.Name]; ok {
				return nil, errors.Errorf("service '%s' is defined multiple times", v.Name)
			}
			v.Directory = d
			services[v.Name] = v
		}
	}
	return services, nil
}

// Add adds a Docker Compose file for the reader to read.
func (r *Reader) Add(path string) {
	r.files = append(r.files, path)
}

// NewReader creates a new Reader.
func NewReader() *Reader {
	return &Reader{}
}
