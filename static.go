package progress

import (
	"fmt"
	"os"
	"strings"

	terminal "github.com/codemodify/systemkit-terminal"
	progress "github.com/codemodify/systemkit-terminal-progress"
)

// static -
type static struct {
	config progress.Config

	stopChannel     chan bool
	stopWithSuccess bool
	finishedChannel chan bool

	lastPrintLen int

	theTerminal *terminal.Terminal
}

// NewStaticWithConfig -
func NewStaticWithConfig(config progress.Config) progress.Renderer {

	// 1. set defaults
	if config.Writer == nil {
		config.Writer = os.Stdout
	}

	// 2.
	return &static{
		config: config,

		stopChannel:     make(chan bool),
		stopWithSuccess: true,
		finishedChannel: make(chan bool),

		lastPrintLen: 0,

		theTerminal: terminal.NewTerminal(config.Writer),
	}
}

// NewStatic -
func NewStatic(args ...string) progress.Renderer {
	return NewStaticWithConfig(*progress.NewDefaultConfig(args...))
}

// Run -
func (thisRef *static) Run() {
	go thisRef.drawLineInLoop()
}

// Success -
func (thisRef *static) Success() {
	thisRef.stop(true)
}

// Fail -
func (thisRef *static) Fail() {
	thisRef.stop(false)
}

func (thisRef *static) stop(success bool) {
	thisRef.stopWithSuccess = success
	thisRef.stopChannel <- true
	close(thisRef.stopChannel)

	<-thisRef.finishedChannel
}

func (thisRef *static) drawLine(char string) (int, error) {
	return fmt.Fprintf(thisRef.config.Writer, "%s%s%s%s", thisRef.config.Prefix, char, thisRef.config.Suffix, thisRef.config.ProgressMessage)
}

func (thisRef *static) drawOperationProgressLine() {
	if err := thisRef.eraseLine(); err != nil {
		return
	}

	n, err := thisRef.drawLine(thisRef.config.ProgressGlyphs[0])
	if err != nil {
		return
	}

	thisRef.lastPrintLen = n
}

func (thisRef *static) drawOperationStatusLine() {
	status := thisRef.config.SuccessGlyph
	if !thisRef.stopWithSuccess {
		status = thisRef.config.FailGlyph
	}

	if err := thisRef.eraseLine(); err != nil {
		return
	}

	if _, err := thisRef.drawLine(status); err != nil {
		return
	}

	fmt.Fprintf(thisRef.config.Writer, "\n")

	thisRef.lastPrintLen = 0
}

func (thisRef *static) drawLineInLoop() {
	if thisRef.config.HideCursor {
		thisRef.theTerminal.CursorHide()
	}

	thisRef.drawOperationProgressLine()

	<-thisRef.stopChannel

	thisRef.drawOperationStatusLine()

	if thisRef.config.HideCursor {
		thisRef.theTerminal.CursorShow()
	}

	thisRef.finishedChannel <- true
}

func (thisRef *static) eraseLine() error {
	_, err := fmt.Fprint(thisRef.config.Writer, "\r"+strings.Repeat(" ", thisRef.lastPrintLen)+"\r")
	return err
}
