package business

import (
	padapter "docker-compose-watcher/internal/provider/adapter"
	"docker-compose-watcher/internal/provider/translator"
	"docker-compose-watcher/internal/rlistener"
	rfsnotify "docker-compose-watcher/internal/rlistener/watcher/fsnotify"
	"docker-compose-watcher/pkg/chanthrottler"
	"docker-compose-watcher/pkg/dockercompose"
	"docker-compose-watcher/pkg/provider"
	pfsnotify "docker-compose-watcher/pkg/provider/watcher/fsnotify"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const throttleDuration = 500 * time.Millisecond

// ComposeController controls compose.
type ComposeController struct {
	p   *provider.Provider
	l   *rlistener.Listener
	cmd *dockercompose.Commander
	exe *exec.Cmd
	rch <-chan provider.ReaderValueWithError
}

func (c *ComposeController) rebuildAndRestart() error {
	if c.exe != nil {
		if err := c.exe.Process.Signal(os.Interrupt); err != nil {
			return err
		}
	}
	c.exe = c.cmd.Build(dockercompose.BuildOptions{})
	c.exe.Stdout = os.Stdout
	c.exe.Stderr = os.Stderr
	if err := c.exe.Run(); err != nil {
		return err
	}
	c.exe = c.cmd.Up(dockercompose.UpOptions{})
	c.exe.Stdout = os.Stdout
	c.exe.Stderr = os.Stderr
	return c.exe.Start()
}

func (c *ComposeController) servicesUpdated(services map[string]translator.WatchedService) error {
	var err error
	if err := c.l.Close(); err != nil {
		return err
	}
	c.l, err = rlistener.New(rfsnotify.New)
	if err != nil {
		return err
	}
	for _, v := range services {
		if v.Path == "" {
			continue
		}
		if err := c.l.AddDir(filepath.Join(v.Directory, v.Path)); err != nil {
			return err
		}
	}
	return c.rebuildAndRestart()
}

// Run runs the compose controller execution loop.
func (c *ComposeController) Run() error {
	c.p.Sync()
	for {
		select {
		case vi, ok := <-chanthrottler.Throttle(throttleDuration, c.l.Channel()):
			if !ok {
				break
			}
			v := vi.(rlistener.ListenerMsg)
			if v.Error != nil {
				return v.Error
			}
			if err := c.rebuildAndRestart(); err != nil {
				return err
			}
		case vi, ok := <-chanthrottler.Throttle(throttleDuration, c.rch):
			if !ok {
				break
			}
			v := vi.(provider.ReaderValueWithError)
			if v.Error != nil {
				return v.Error
			}
			if err := c.servicesUpdated(v.Value.(map[string]translator.WatchedService)); err != nil {
				return err
			}
		}
	}
}

// Close cleans up the controller.
func (c *ComposeController) Close() error {
	return c.p.Close()
}

// NewComposeController creates a new compose controller.
func NewComposeController(composePaths ...string) (*ComposeController, error) {
	x, err := provider.New(padapter.NewServiceReader, pfsnotify.New)
	if err != nil {
		return nil, err
	}
	l, err := rlistener.New(rfsnotify.New)
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
		l:   l,
	}, nil
}
