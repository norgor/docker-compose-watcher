package serviceprovider

import (
	"io"
	"strconv"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Service struct {
	Name string
}

type BuiltService struct {
	Service
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

func readServices(reader io.Reader) ([]BuiltService, error) {
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
			Service: Service{Name: serviceName},
			Path:    path,
		})
	}
	return services, nil
}
