package termios

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// Termios represents the state of a terminal at a point in time.
type Termios struct {
	fd    int
	state *unix.Termios
}

// EnableRawMode tells the terminal described by fd to read raw input, writing all key
// presses directly to the file without interpretation. It returns the original state of the
// terminal, which should be restored before the program exits.
func EnableRawMode(fd int) (*Termios, error) {
	// Load the initial terminal state from the OS.
	initialState, err := unix.IoctlGetTermios(fd, IoctlGetTermiosReq)
	if err != nil {
		return nil, fmt.Errorf("get termios for FD %d: %w", fd, err)
	}

	raw := *initialState
	// Disable the flags required to put the terminal into raw mode. See `man termios` for more.
	//
	// Local (miscellaneous) flags. From left to right, disable:
	// - echoing;
	// - canonical mode, allowing us to read byte by byte instead of line by line;
	// - implementation-defined input processing (Ctrl-V), which, on Darwin, awaits another keypress
	//   before interpreting it as an integer literal (e.g. Ctrl-V Ctrl-C would output 3);
	// - interrupts via Ctrl-C and Ctrl-Z.
	raw.Lflag &^= unix.ECHO | unix.ICANON | unix.IEXTEN | unix.ISIG
	// Input flags. From left to right, disable:
	// - mapping of BREAK to SIGINTR
	// - mapping of CR to NL
	// - checking of parity errors (not applicable to modern terminal emulators)
	// - output flow control (Ctrl-S suspends flow of data from the terminal and Ctrl-Q resumes it)
	raw.Iflag &^= unix.BRKINT | unix.ICRNL | unix.INPCK | unix.IXON
	// Output flags. Disable:
	// - automatic conversion of \n to \r\n
	raw.Oflag &^= unix.OPOST
	// Control modes. Apply mask CS8 to set the character size to 8 bits.
	raw.Cflag |= unix.CS8

	if err := unix.IoctlSetTermios(fd, IoctlSetTermiosReq, &raw); err != nil {
		return nil, fmt.Errorf("set termios for FD %d: %w", fd, err)
	}

	return &Termios{fd: fd, state: initialState}, nil
}

// DisableRawMode restores the terminal represented by t to its original state.
func (t *Termios) DisableRawMode() error {
	if err := unix.IoctlSetTermios(t.fd, IoctlSetTermiosReq, t.state); err != nil {
		return fmt.Errorf("disable raw mode: %w", err)
	}
	return nil
}
