//go:build !windows && !plan9
// +build !windows,!plan9

package geamon

import "syscall"

func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setsid: true,
	}
}
