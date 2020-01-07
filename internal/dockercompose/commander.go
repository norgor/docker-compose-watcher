package dockercompose

import (
	"fmt"
	"reflect"

	"os/exec"
)

// LogLevel constants
const (
	LogDebug    = LogLevel("DEBUG")
	LogInfo     = LogLevel("INFO")
	LogWarning  = LogLevel("WARNING")
	LogError    = LogLevel("ERROR")
	LogCritical = LogLevel("CRITICAL")
)

const (
	composeExecutable = "docker-compose"
	buildCmd          = "build"
	upCmd             = "up"
	tagName           = "compose-option"
)

// LogLevel type for describing level of logging
type LogLevel string

// CommanderOptions specifies the global flags for docker-compose
type CommanderOptions struct {
	Files             []string `compose-option:"-f"`
	ProjectName       string   `compose-option:"-p"`
	Verbose           bool     `compose-option:"--verbose"`
	LogLevel          LogLevel `compose-option:"--log-level"`
	NoAnsi            bool     `compose-option:"--no-ansi"`
	Version           bool     `compose-option:"--version"`
	Host              string   `compose-option:"-H"`
	TLS               bool     `compose-option:"--tls"`
	TLSCACert         string   `compose-option:"--tlscacert"`
	TLSCert           string   `compose-option:"--tlscert"`
	TLSKey            string   `compose-option:"--tlskey"`
	TLSVerify         bool     `compose-option:"--tlsverify"`
	SkipHostnameCheck bool     `compose-option:"--skip-hostname-check"`
	ProjectDirectory  string   `compose-option:"--project-directory"`
	Compatibility     bool     `compose-option:"--compatibility"`
}

// Commander is used to prepare commands for docker-compose
type Commander struct {
	opt     CommanderOptions
	optArgs []string
}

// BuildOptions are used to specify options (flags) for the 'docker-compose build' command
type BuildOptions struct {
	Compress  bool              `compose-option:"--compress"`
	ForceRM   bool              `compose-option:"--force-rm"`
	NoCache   bool              `compose-option:"--no-cache"`
	Pull      bool              `compose-option:"--pull"`
	Memory    int               `compose-option:"-m"`
	BuildArgs map[string]string `compose-option:"--build-arg"`
	Parallel  bool              `compose-option:"--parallel"`
}

// UpOptions are used to specify options (flags) for the 'docker-compose up' command
type UpOptions struct {
	Detach               bool           `compose-option:"-d"`
	NoColor              bool           `compose-option:"--no-color"`
	QuietPull            bool           `compose-option:"--quiet-pull"`
	NoDeps               bool           `compose-option:"--no-deps"`
	ForceRecreate        bool           `compose-option:"--force-recreate"`
	AlwaysCreateDeps     bool           `compose-option:"--always-recreate-deps"`
	NoRecreate           bool           `compose-option:"--no-recreate"`
	NoBuild              bool           `compose-option:"--no-build"`
	NoStart              bool           `compose-option:"--no-start"`
	Build                bool           `compose-option:"--build"`
	AbortOnContainerExit bool           `compose-option:"--abort-on-container-exit"`
	Timeout              int            `compose-option:"-t"`
	RenewAnonVolumes     bool           `compose-option:"-V"`
	RemoveOrphans        bool           `compose-option:"--remove-orphans"`
	ExitCodeFrom         string         `compose-option:"--exit-code-from"`
	Scale                map[string]int `compose-option:"--scale"`
}

func taggedValueToArgs(tag string, value interface{}, ignoreZero bool) (args []string) {
	to := reflect.TypeOf(value)
	vo := reflect.ValueOf(value)

	if !ignoreZero && vo.IsZero() {
		return nil
	}

	switch to.Kind() {
	case reflect.Bool:
		args = append(args, tag)
	case reflect.Int:
		args = append(args, tag, fmt.Sprintf("%#v", vo.Int()))
	case reflect.String:
		args = append(args, tag, fmt.Sprintf("%#v", vo.String()))
	case reflect.Slice:
		for i := 0; i < vo.Len(); i++ {
			v := vo.Index(i).Interface()
			args = append(args, taggedValueToArgs(tag, v, true)...)
		}
	case reflect.Map:
		if to.Key().Kind() != reflect.String {
			panic("unable to argumentize maps with non-string key")
		}
		iter := vo.MapRange()
		for iter.Next() {
			args = append(args, tag, fmt.Sprintf("%#v=%#v", iter.Key(), iter.Value()))
		}
	default:
		panic(fmt.Errorf("unable to argumentize kind %v", to.Kind()))
	}

	return args
}

func optionsToArgs(opt interface{}) (args []string) {
	if opt == nil {
		return nil
	}

	t := reflect.TypeOf(opt)
	vo := reflect.ValueOf(opt)
	for i := 0; i < t.NumField(); i++ {
		v := vo.Field(i)
		f := t.Field(i)
		vi := v.Interface()

		tag := f.Tag.Get(tagName)
		if tag == "" {
			continue
		}

		args = append(args, taggedValueToArgs(tag, vi, false)...)
	}

	return args
}

// Command returns a 'docker-compose <cmd>' command with the specified arguments.
func (e *Commander) Command(cmd string, arg ...string) *exec.Cmd {
	var args []string
	args = append(args, e.optArgs...)
	args = append(args, cmd)
	args = append(args, arg...)
	return exec.Command(composeExecutable, args...)
}

func (e *Commander) commandWithOptions(cmd string, opt interface{}) *exec.Cmd {
	return e.Command(cmd, optionsToArgs(opt)...)
}

// Build returns a 'docker-compose build' command with the specified options.
func (e *Commander) Build(opt BuildOptions) *exec.Cmd {
	return e.commandWithOptions(buildCmd, opt)
}

// Up returns a 'docker-compose up' command with the specified options.
func (e *Commander) Up(opt UpOptions) *exec.Cmd {
	return e.commandWithOptions(upCmd, opt)
}

// NewCommander creates a new commander instance with the specified options (global flags),
// which will be used when executing commands.
func NewCommander(opt CommanderOptions) *Commander {
	return &Commander{
		opt:     opt,
		optArgs: optionsToArgs(opt),
	}
}
