package service

import (
	"io"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type BuiltService struct {
	Name string
	Path string
}

type composeService struct {
	Build interface{} `yaml:build`
}

type compose struct {
	Version  string `yaml:version`
	Services map[string]composeService
}

type version struct {
	Major int
	Minor int
}

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

func getServiceBuildPath(service *composeService) (string, error) {
	switch build := service.Build.(type) {
	case string:
		return build, nil
	case map[interface{}]interface{}:
		ctx, ok := build["context"].(string)
		if !ok {
			return "", errors.New("service.build.context was not of type string")
		}
		return ctx, nil
	}
	return "", errors.New("service.build was of invalid type")
}

func readFromCompose(reader io.Reader) ([]BuiltService, error) {
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

func transformServices(compose *compose) ([]BuiltService, error) {
	var services []BuiltService
	for serviceName, service := range compose.Services {
		if service.Build == nil {
			continue
		}
		path, err := getServiceBuildPath(&service)
		if err != nil {
			return nil, errors.Wrapf(err, "error while transforming service '%s'", serviceName)
		}
		services = append(services, BuiltService{
			Name: serviceName,
			Path: path,
		})
	}
	return services, nil
}

func (r *Reader) Read() (map[string]BuiltService, error) {
	services := make(map[string]BuiltService)
	for _, file := range r.files {
		file, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		s, err := readFromCompose(file)
		if err != nil {
			return nil, err
		}
		for _, v := range s {
			if _, ok := services[v.Name]; ok {
				return nil, errors.Errorf("service '%s' is defined multiple times")
			}
			services[v.Name] = v
		}
	}
	return services, nil
}

func (r *Reader) AddCompose(path string) {
	sr.files = append(r.files, path)
}

func NewServiceReader() *Reader {
	return &Reader{}
}
