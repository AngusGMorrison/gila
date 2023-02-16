//go:build aix && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris && !zos

package termios

func init() {
	panic(fmt.Sprintf("Platform %s/%s is not supported", runtime.GOOS, runtime.GOARCH))
}
