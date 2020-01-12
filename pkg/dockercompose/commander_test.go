package dockercompose

import (
	"reflect"
	"testing"
)

func TestCommander_Command(t *testing.T) {
	type cargs struct {
		opt CommanderOptions
	}
	type args struct {
		cmd string
		arg []string
	}

	tests := []struct {
		name        string
		cargs       cargs
		args        args
		wantCmdArgs []string
	}{
		{
			name: "passes the specified flags correctly",
			cargs: cargs{CommanderOptions{
				Files:             []string{"foo", "bar"},
				ProjectName:       "baz",
				Verbose:           true,
				LogLevel:          LogDebug,
				NoAnsi:            true,
				Version:           true,
				Host:              "boo",
				TLS:               true,
				TLSCACert:         "bazinga",
				TLSCert:           "foocert",
				TLSKey:            "fookey",
				TLSVerify:         true,
				SkipHostnameCheck: true,
				ProjectDirectory:  "foodir",
				Compatibility:     true,
			}},
			args: args{"foo-command", nil},
			wantCmdArgs: []string{
				"docker-compose",
				"-f", `foo`,
				"-f", `bar`,
				"-p", `baz`,
				"--verbose",
				"--log-level", `DEBUG`,
				"--no-ansi",
				"--version",
				"-H", `boo`,
				"--tls",
				"--tlscacert", `bazinga`,
				"--tlscert", `foocert`,
				"--tlskey", `fookey`,
				"--tlsverify",
				"--skip-hostname-check",
				"--project-directory", `foodir`,
				"--compatibility",
				"foo-command",
			},
		},
		{
			name:  "does not pass the unspecified flags",
			cargs: cargs{CommanderOptions{}},
			args:  args{"foo-command", nil},
			wantCmdArgs: []string{
				"docker-compose", "foo-command",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewCommander(tt.cargs.opt)
			if got := e.Command(tt.args.cmd, tt.args.arg...); !reflect.DeepEqual(got.Args, tt.wantCmdArgs) {
				t.Errorf("Commander.Command() = %v, want %v", got.Args, tt.wantCmdArgs)
			}
		})
	}
}

func TestCommander_Up(t *testing.T) {
	type args struct {
		opt UpOptions
	}
	tests := []struct {
		name        string
		args        args
		wantCmdArgs []string
	}{
		{
			name: "passes the specified flags correctly",
			args: args{UpOptions{
				Detach:               true,
				NoColor:              true,
				QuietPull:            true,
				NoDeps:               true,
				ForceRecreate:        true,
				AlwaysCreateDeps:     true,
				NoRecreate:           true,
				NoBuild:              true,
				NoStart:              true,
				Build:                true,
				AbortOnContainerExit: true,
				Timeout:              64,
				RenewAnonVolumes:     true,
				RemoveOrphans:        true,
				ExitCodeFrom:         "foo",
				Scale:                map[string]int{"foo": 128},
			}},
			wantCmdArgs: []string{
				"docker-compose", "up",
				"-d",
				"--no-color",
				"--quiet-pull",
				"--no-deps",
				"--force-recreate",
				"--always-recreate-deps",
				"--no-recreate",
				"--no-build",
				"--no-start",
				"--build",
				"--abort-on-container-exit",
				"-t", "64",
				"-V",
				"--remove-orphans",
				"--exit-code-from", `foo`,
				"--scale", `"foo"=128`,
			},
		},
		{
			name: "does not pass the unspecified flags",
			args: args{UpOptions{}},
			wantCmdArgs: []string{
				"docker-compose", "up",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewCommander(CommanderOptions{})
			if got := e.Up(tt.args.opt); !reflect.DeepEqual(got.Args, tt.wantCmdArgs) {
				t.Errorf("Commander.Up() = %v, want %v", got.Args, tt.wantCmdArgs)
			}
		})
	}
}

func TestCommander_Build(t *testing.T) {
	type args struct {
		opt BuildOptions
	}
	tests := []struct {
		name        string
		args        args
		wantCmdArgs []string
	}{
		{
			name: "passes the specified flags correctly",
			args: args{BuildOptions{
				Compress:  true,
				ForceRM:   true,
				NoCache:   true,
				Pull:      true,
				Memory:    64,
				BuildArgs: map[string]string{"a": "b"},
				Parallel:  true,
			}},
			wantCmdArgs: []string{
				"docker-compose", "build",
				"--compress",
				"--force-rm",
				"--no-cache",
				"--pull",
				"-m", "64",
				"--build-arg", `"a"="b"`,
				"--parallel",
			},
		},
		{
			name: "does not pass the unspecified flags",
			args: args{BuildOptions{}},
			wantCmdArgs: []string{
				"docker-compose", "build",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewCommander(CommanderOptions{})
			if got := e.Build(tt.args.opt); !reflect.DeepEqual(got.Args, tt.wantCmdArgs) {
				t.Errorf("Commander.Build() = %v, want %v", got.Args, tt.wantCmdArgs)
			}
		})
	}
}
