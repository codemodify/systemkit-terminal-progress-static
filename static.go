package progress

import (
	"fmt"
	"os"
	"strings"

	terminal "github.com/codemodify/systemkit-terminal"
	progress "github.com/codemodify/systemkit-terminal-progress"
)

// Static -
type Static struct {
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
	return &Static{
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
	progressMessage := ""
	successMessage := ""
	failMessage := ""

	if len(args) > 0 {
		progressMessage = args[0]
	}

	if len(args) > 1 {
		successMessage = args[1]
	} else {
		successMessage = progressMessage
	}

	if len(args) > 2 {
		failMessage = args[2]
	} else {
		failMessage = progressMessage
	}

	return NewStaticWithConfig(progress.Config{
		Prefix:          "[",
		ProgressGlyphs:  []string{string('\u25B6')}, // u00BB - double arrow, u25B6 - play
		Suffix:          "] ",
		ProgressMessage: progressMessage,
		SuccessGlyph:    string('\u2713'), // check mark
		SuccessMessage:  successMessage,
		FailGlyph:       string('\u00D7'), // middle cross
		FailMessage:     failMessage,
		Writer:          os.Stdout,
		HideCursor:      true,
	})
}

// Run -
func (thisRef *Static) Run() {
	go thisRef.drawLineInLoop()
}

// Success -
func (thisRef *Static) Success() {
	thisRef.stop(true)
}

// Fail -
func (thisRef *Static) Fail() {
	thisRef.stop(false)
}

func (thisRef *Static) stop(success bool) {
	thisRef.stopWithSuccess = success
	thisRef.stopChannel <- true
	close(thisRef.stopChannel)

	<-thisRef.finishedChannel
}

func (thisRef *Static) drawLine(char string) (int, error) {
	return fmt.Fprintf(thisRef.config.Writer, "%s%s%s%s", thisRef.config.Prefix, char, thisRef.config.Suffix, thisRef.config.ProgressMessage)
}

func (thisRef *Static) drawOperationProgressLine() {
	if err := thisRef.eraseLine(); err != nil {
		return
	}

	n, err := thisRef.drawLine(thisRef.config.ProgressGlyphs[0])
	if err != nil {
		return
	}

	thisRef.lastPrintLen = n
}

func (thisRef *Static) drawOperationStatusLine() {
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

func (thisRef *Static) drawLineInLoop() {
	if thisRef.config.HideCursor {
		thisRef.theTerminal.HideCursor()
	}

	thisRef.drawOperationProgressLine()

	<-thisRef.stopChannel

	thisRef.drawOperationStatusLine()

	if thisRef.config.HideCursor {
		thisRef.theTerminal.ShowCursor()
	}

	thisRef.finishedChannel <- true
}

func (thisRef *Static) eraseLine() error {
	_, err := fmt.Fprint(thisRef.config.Writer, "\r"+strings.Repeat(" ", thisRef.lastPrintLen)+"\r")
	return err
}
