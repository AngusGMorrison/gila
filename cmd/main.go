package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/angusgmorrison/gila/bufio"
	"github.com/angusgmorrison/gila/editor"
	"github.com/angusgmorrison/gila/escseq"
	"github.com/angusgmorrison/gila/renderer"
	"golang.org/x/term"
)

const (
	logFile = "editor.log"
	name    = "Gila editor"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}

func run() (err error) {
	var filepath string
	if len(os.Args) > 1 {
		filepath = os.Args[1]
	}

	// Enable terminal raw mode to process each keypress as it happens.
	initialTermState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("enable terminal raw mode: %w", err)
	}
	defer func() { err = term.Restore(int(os.Stdin.Fd()), initialTermState) }()
	// In raw mode, the cursor won't return to the start of the next line after
	// the terminal echoes the command used to run the program, so we force the
	// line feed.
	fmt.Print("\r")

	keyReader := bufio.NewKeyReader(os.Stdin, escseq.MaxLenBytes)
	terminalWriter := bufio.NewTerminalWriter(os.Stdout)
	info, _ := debug.ReadBuildInfo()
	w, h, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("get terminal size: %w", err)
	}
	renderer := renderer.New(
		name,
		info.Main.Version,
		terminalWriter,
		renderer.Screen{
			Width:  w,
			Height: h,
		},
	)

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer f.Close()
	logger := log.New(f, "", log.LstdFlags|log.Lshortfile)

	ed := editor.New(
		keyReader,
		renderer,
		editor.Config{
			Width:  w,
			Height: h,
		},
		logger,
	)
	return ed.Run(filepath)
}
