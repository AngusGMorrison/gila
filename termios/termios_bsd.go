//go:build darwin || dragonfly || freebsd || openbsd || netbsd

package termios

import "golang.org/x/sys/unix"

const (
	IoctlGetTermiosReq = unix.TIOCGETA
	IoctlSetTermiosReq = unix.TIOCSETA
)
