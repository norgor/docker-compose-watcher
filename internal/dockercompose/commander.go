package dockercompose

import (
	"fmt"
	"reflect"
	"strconv"

	"os/exec"
)

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
	tagName           = "dcf"
)

type LogLevel string

type CommanderOptions struct {
	Files             []string `dcf:"-f"`
	ProjectName       string   `dcf:"-p"`
	Verbose           bool     `dcf:"--verbose"`
	LogLevel          LogLevel `dcf:"--log-level"`
	NoAnsi            bool     `dcf:"--no-ansi"`
	Version           bool     `dcf:"--version"`
	Host              string   `dcf:"-H"`
	TLS               bool     `dcf:"--tls"`
	TLSCACert         string   `dcf:"--tlscacert"`
	TLSCert           string   `dcf:"--tlscert"`
	TLSKey            string   `dcf:"--tlskey"`
	TLSVerify         bool     `dcf:"--tlsverify"`
	SkipHostnameCheck bool     `dcf:"--skip-hostname-check"`
	ProjectDirectory  string   `dcf:"--project-directory"`
	Compatibility     bool     `dcf:"--compatibility"`
}

type Commander struct {
	ctx     CommanderOptions
	ctxArgs []string
}

type BuildOptions struct {
	Compress  bool              `dcf:"--compress"`
	ForceRM   bool              `dcf:"--force-rm"`
	NoCache   bool              `dcf:"--no-cache"`
	Pull      bool              `dcf:"--pull"`
	Memory    int               `dcf:"-m"`
	BuildArgs map[string]string `dcf:"--build-arg"`
	Parallel  bool              `dcf:"--parallel"`
}

type UpOptions struct {
	Detach               bool           `dcf:"-d"`
	NoColor              bool           `dcf:"--no-color"`
	QuietPull            bool           `dcf:"--quiet-pull"`
	NoDeps               bool           `dcf:"--no-deps"`
	ForceRecreate        bool           `dcf:"--force-recreate"`
	AlwaysCreateDeps     bool           `dcf:"--always-recreate-deps"`
	NoRecreate           bool           `dcf:"--no-recreate"`
	NoBuild              bool           `dcf:"--no-build"`
	NoStart              bool           `dcf:"--no-start"`
	Build                bool           `dcf:"--build"`
	AbortOnContainerExit bool           `dcf:"--abort-on-container-exit"`
	Timeout              int            `dcf:"-t"`
	RenewAnonVolumes     bool           `dcf:"-V"`
	RemoveOrphans        bool           `dcf:"--remove-orphans"`
	ExitCodeFrom         string         `dcf:"--exit-code-from"`
	Scale                map[string]int `dcf:"--scale"`
}

func taggedValueToArgs(tag string, value interface{}, ignoreZero bool) (args []string) {
	to := reflect.TypeOf(value)
	vo := reflect.ValueOf(value)

	if ignoreZero && value == reflect.Zero(to).Interface() {
		return nil
	}

	switch to.Kind() {
	case reflect.Bool:
		args = append(args, tag)
	case reflect.Int:
		args = append(args, tag, strconv.Itoa(value.(int)))
	case reflect.String:
		args = append(args, tag, string(value.(string)))
	case reflect.Array:
		for i := 0; i < vo.Len(); i++ {
			v := vo.Slice(i, i+1)
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

func optionsToArgs(ctx interface{}) (args []string) {
	t := reflect.TypeOf(ctx)
	for i := 0; i < t.NumField(); i++ {
		v := reflect.ValueOf(t).Field(i)
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

func (e *Commander) composeCommand(cmd string, arg ...string) *exec.Cmd {
	var args []string
	args = append(args, e.ctxArgs...)
	args = append(args, arg...)
	return exec.Command(composeExecutable, args...)
}

func (e *Commander) composeCmdWithOptions(cmd string, opt interface{}) *exec.Cmd {
	return e.composeCommand(cmd, optionsToArgs(opt)...)
}

func (e *Commander) Build(opt BuildOptions) *exec.Cmd {
	return e.composeCmdWithOptions(buildCmd, opt)
}

func (e *Commander) Up(opt UpOptions) *exec.Cmd {
	return e.composeCmdWithOptions(buildCmd, opt)
}

func NewExecutor(ctx CommanderOptions) *Commander {
	return &Commander{
		ctx:     ctx,
		ctxArgs: optionsToArgs(ctx),
	}
}
