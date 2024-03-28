package geamon

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/erikdubbelboer/gspt"
)

var _INIT_ENV_KEY = "_GEAMON_INIT"

type Geamon struct {
	Stdout       io.Writer // log output, must
	Stderr       io.Writer // Error log output, must
	PidFile      string    // pid file; Only one instance can be started for the same file
	ProcessTitle string    // Process title
}

func (g *Geamon) _init() error {
	cmd, err := fork(g.Stdout, g.Stderr, []string{fmt.Sprintf("%s=%d", _INIT_ENV_KEY, 1)})
	if err != nil {
		return err
	}

	g.Stdout.Write([]byte(fmt.Sprintf("exec: %s\n", strings.Join(cmd.Args, " "))))
	g.Stdout.Write([]byte("background mode initializing...\n"))
	os.Exit(0)
	return nil
}

func (g *Geamon) _getPifFile() (*os.File, error) {
	if g.PidFile != "" {
		filename, err := filepath.Abs(g.PidFile)
		if err != nil {
			return nil, err
		}
		g.PidFile = filename
		if err := os.MkdirAll(filepath.Dir(g.PidFile), 0755); err != nil {
			return nil, err
		}
		file, err := os.OpenFile(g.PidFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err != nil {
			return nil, fmt.Errorf("an instance is already running at: %s", g.PidFile)
		}
		return file, nil
	}
	return nil, nil
}

func (g *Geamon) _writePidFile() {
	if g.PidFile != "" {
		file, err := g._getPifFile()
		if err != nil {
			g.Stderr.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
			os.Exit(1)
		}
		pid := strconv.Itoa(os.Getpid())
		_, err = file.WriteString(pid + "\n")
		if err != nil {
			g.Stderr.Write([]byte(fmt.Sprintf("could not write pid to lockfile: %s\n", err.Error())))
			os.Exit(1)
		}
		g.Stdout.Write([]byte(fmt.Sprintf("pid file: %s\n", g.PidFile)))
	}
}

func (g *Geamon) ReleasePidFile() {
	if g.PidFile != "" {
		file, err := os.OpenFile(g.PidFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			g.Stderr.Write([]byte(fmt.Sprintf("open pid file error: %s\n", err.Error())))
			os.Exit(1)
		}
		syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		file.Close()
		os.Remove(g.PidFile)
	}
}

func (g *Geamon) IsBG() bool {
	return os.Getenv(_INIT_ENV_KEY) == "1"
}

// start deamon
func (g *Geamon) Run() error {
	if !g.IsBG() {
		if _, err := g._getPifFile(); err != nil {
			return err
		}
		return g._init()
	}
	g.Stdout.Write([]byte(fmt.Sprintf("running pid: %d\n", os.Getpid())))
	title := strings.Trim(g.ProcessTitle, " ")
	if title != "" {
		gspt.SetProcTitle(title)
	}
	g._writePidFile()
	g.Stdout.Write([]byte("successfully started\n"))
	return nil
}
