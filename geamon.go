package geamon

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/erikdubbelboer/gspt"
)

type PType int

const (
	PT_INIT          PType = 1
	PT_DEAMON        PType = 2
	PT_WORKER        PType = 3
	PROCESS_TYPE_KEY       = "GEAMON_PROCESS_TYPE_KEY"
)

func (p PType) IsInit() bool {
	return p == PT_INIT || p == 0
}

func (p PType) IsDeamon() bool {
	return p == PT_DEAMON
}

func (p PType) IsWorker() bool {
	return p == PT_WORKER || (!p.IsInit() && !p.IsDeamon())
}

func ProcessType() PType {
	i, _ := strconv.Atoi(os.Getenv(PROCESS_TYPE_KEY))
	return PType(i)
}

func fork(stdout io.Writer, stderr io.Writer, env []string) (*exec.Cmd, error) {
	cmd := &exec.Cmd{
		Path: os.Args[0],
		Args: os.Args,
	}

	envs := os.Environ()
	newEnvs := []string{}
	for _, v := range envs {
		if !strings.Contains(v, fmt.Sprintf("%s=", PROCESS_TYPE_KEY)) {
			newEnvs = append(newEnvs, v)
		}
	}
	newEnvs = append(newEnvs, env...)
	cmd.Env = newEnvs
	cmd.SysProcAttr = sysProcAttr()
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return cmd, nil
}

type Geamon struct {
	Stdout     io.Writer // log output, must
	Stderr     io.Writer // Error log output, must
	PidFile    string    // pid file; Only one instance can be started for the same file
	MaxRestart uint      // Maximum number of restarts, 0 unlimited
	DeamonName string    // Process title
}

func (g *Geamon) _init() error {
	_, err := fork(g.Stdout, g.Stderr, []string{fmt.Sprintf("%s=%d", PROCESS_TYPE_KEY, PT_DEAMON)})
	if err != nil {
		return err
	}

	g.Stdout.Write([]byte("Daemon initializing...\n"))
	os.Exit(0)
	return nil
}

func (g *Geamon) _deamon() {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
	maxRestart := g.MaxRestart
	pid := os.Getpid()
	g.Stdout.Write([]byte(fmt.Sprintf("Daemon started successfully, pid: %d\n", pid)))

	if g.DeamonName != "" {
		gspt.SetProcTitle(g.DeamonName)
	}
	for {
		cmd, err := fork(g.Stdout, g.Stderr, []string{fmt.Sprintf("%s=%d", PROCESS_TYPE_KEY, PT_WORKER)})
		if err != nil {
			panic(err)
		}
		g.Stdout.Write([]byte(fmt.Sprintf("Started worker(%d)\n", cmd.Process.Pid)))
		wch := make(chan bool, 1)
		go func() {
			if err := cmd.Wait(); err != nil {
				g.Stderr.Write([]byte(fmt.Sprintf("%d Worker exited(%d): %s\n", cmd.Process.Pid, cmd.ProcessState.ExitCode(), err.Error())))
				wch <- true
			} else {
				g.Stdout.Write([]byte(fmt.Sprintf("%d Worker exited(%d)\n", cmd.Process.Pid, cmd.ProcessState.ExitCode())))
			}
		}()
		select {
		case <-wch:
			if maxRestart >= 1 {
				maxRestart--
			}
			if maxRestart == 0 && g.MaxRestart != 0 {
				g.Stdout.Write([]byte(fmt.Sprintf("Exceeded the maximum number of restarts(%d), exiting\n", g.MaxRestart)))
				g.Stdout.Write([]byte("Deamon exited\n"))
				g._releasePidFile()
				os.Exit(0)
			}
		case sig := <-exit:
			g.Stdout.Write([]byte(fmt.Sprintf("Deamon received exit signal(%d)\n", sig)))
			err := cmd.Process.Kill()
			if err != nil {
				g.Stderr.Write([]byte(fmt.Sprintf("Kill worker(%d) error: %s\n", cmd.Process.Pid, err.Error())))
			}
			g.Stdout.Write([]byte(fmt.Sprintf("%d Worker exited\n", cmd.Process.Pid)))
			g.Stdout.Write([]byte("Deamon exited\n"))
			g._releasePidFile()
			os.Exit(0)
		}
	}
}

func (g *Geamon) _getPifFile() (*os.File, error) {
	if g.PidFile != "" {
		file, err := os.OpenFile(g.PidFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err != nil {
			return nil, errors.New("another instance is already running")
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
		_, err = file.WriteString(pid)
		if err != nil {
			g.Stderr.Write([]byte(fmt.Sprintf("Could not write PID to lockfile: %s\n", err.Error())))
			os.Exit(1)
		}
		g.Stdout.Write([]byte(fmt.Sprintf("Start with pid file: %s\n", g.PidFile)))
	}
}

func (g *Geamon) _releasePidFile() {
	if g.PidFile != "" {
		file, err := os.OpenFile(g.PidFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			g.Stderr.Write([]byte(fmt.Sprintf("Open pid file error: %s\n", err.Error())))
			os.Exit(1)
		}
		syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		file.Close()
		os.Remove(g.PidFile)
	}
}

// start deamon
func (g *Geamon) Run() error {
	pt := ProcessType()
	if pt.IsInit() {
		if _, err := g._getPifFile(); err != nil {
			return err
		}
		return g._init()
	} else if pt.IsDeamon() {
		g._writePidFile()
		g._deamon()
		return nil
	}
	return nil
}
