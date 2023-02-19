package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/angusgmorrison/gila/bufio"
	"github.com/angusgmorrison/gila/editor"
	"github.com/angusgmorrison/gila/escseq"
	"golang.org/x/term"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}

func run() (err error) {
	// Enable terminal raw mode to process each keypress as it happens.
	initialTermState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("enable terminal raw mode: %w", err)
	}
	defer func() { err = term.Restore(int(os.Stdin.Fd()), initialTermState) }()
	// In raw mode, the cursor won't return to the start of the next line after the terminal echoes
	// the command used to run the program, so we force the line feed.
	fmt.Print("\r")

	w, h, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("get terminal size: %w", err)
	}

	info, _ := debug.ReadBuildInfo()
	config := editor.Config{
		Name:    "Gila editor",
		Version: info.Main.Version,
		Width:   uint(w),
		Height:  uint(h),
	}
	keyReader := bufio.NewKeyReader(os.Stdin, escseq.MaxLenBytes)
	terminalWriter := bufio.NewTerminalWriter(os.Stdout)
	ed := editor.New(keyReader, terminalWriter, config)
	return ed.Run()
}
