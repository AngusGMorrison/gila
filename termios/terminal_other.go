//go:build aix || linux || solaris || zos

package termios

import "golang.org/x/sys/unix"

const (
	IoctlGetTermiosReq = unix.TCGETA
	IoctlSetTermiosReq = unix.TCSETA
)
