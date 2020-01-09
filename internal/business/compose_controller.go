package business

import (
	padapter "docker-compose-watcher/internal/provider/adapter"
	"docker-compose-watcher/internal/provider/translator"
	"docker-compose-watcher/pkg/dockercompose"
	"docker-compose-watcher/pkg/provider"
	"docker-compose-watcher/pkg/provider/watcher/fsnotify"
	"os"
	"os/exec"
)

// ComposeController controls compose.
type ComposeController struct {
	p   *provider.Provider
	cmd *dockercompose.Commander
	exe *exec.Cmd
	rch <-chan provider.ReaderValueWithError
}

func (c *ComposeController) servicesUpdated(services map[string]translator.WatchedService) error {
	// TODO: reload recursive reader
	panic("not implemented")
	if c.exe != nil {
		if err := c.exe.Process.Signal(os.Interrupt); err != nil {
			return err
		}
	}
	if err := c.cmd.Build(dockercompose.BuildOptions{}).Run(); err != nil {
		return err
	}
	c.exe = c.cmd.Up(dockercompose.UpOptions{})
	return c.exe.Start()
}

// Run runs the compose controller execution loop.
func (c *ComposeController) Run() error {
	for {
		select {
		case v, ok := <-c.p.Channel():
			if !ok {
				return
			}
			if v.Error != nil {
				return v.Error
			}
			c.servicesUpdated(v.Value.(map[string]translator.WatchedService))
		}
	}
}

// Close cleans up the controller.
func (c *ComposeController) Close() error {
	return c.p.Close()
}

// NewComposeController creates a new compose controller.
func NewComposeController(composePaths ...string) (*ComposeController, error) {
	x, err := provider.New(padapter.NewServiceReader, fsnotify.New)
	if err != nil {
		return nil, err
	}
	for _, v := range composePaths {
		err := x.Add(v)
		if err != nil {
			x.Close()
			return nil, err
		}
	}
	c := dockercompose.NewCommander(dockercompose.CommanderOptions{
		Files: composePaths,
	})
	r := translator.NewServiceTranslatorChannel(x.Channel())
	return &ComposeController{
		p:   x,
		cmd: c,
		rch: r,
	}, nil
}
