package progress

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	progress "github.com/codemodify/systemkit-terminal-progress"
)

// Static -
type Static struct {
	config progress.Config

	stopChannel     chan bool
	stopWithSuccess bool
	finishedChannel chan bool

	lastPrintLen int
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
		ProgressGlyphs:  []string{"|", "/", "-", "\\"},
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
func (s *Static) Run() {
	go s.drawLineInLoop()
}

// Success -
func (s *Static) Success() {
	s.stop(true)
}

// Fail -
func (s *Static) Fail() {
	s.stop(false)
}

func (s *Static) stop(success bool) {
	s.stopWithSuccess = success
	s.stopChannel <- true
	close(s.stopChannel)

	<-s.finishedChannel
}

func (s *Static) drawLine(char string) (int, error) {
	return fmt.Fprintf(s.config.Writer, "%s%s%s%s", s.config.Prefix, char, s.config.Suffix, s.config.ProgressMessage)
}

func (s *Static) drawOperationProgressLine() {
	if err := s.eraseLine(); err != nil {
		return
	}

	n, err := s.drawLine(s.config.ProgressGlyphs[0])
	if err != nil {
		return
	}

	s.lastPrintLen = n
}

func (s *Static) drawOperationStatusLine() {
	status := s.config.SuccessGlyph
	if !s.stopWithSuccess {
		status = s.config.FailGlyph
	}

	if err := s.eraseLine(); err != nil {
		return
	}

	if _, err := s.drawLine(status); err != nil {
		return
	}

	fmt.Fprintf(s.config.Writer, "\n")

	s.lastPrintLen = 0
}

func (s *Static) drawLineInLoop() {
	s.hideCursor()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				s.drawOperationProgressLine()

			case <-s.stopChannel:
				ticker.Stop()
				return
			}
		}
	}()

	// for stop aka Success/Fail
	wg.Wait()

	s.drawOperationStatusLine()

	s.unhideCursor()

	s.finishedChannel <- true
}

func (s *Static) eraseLine() error {
	_, err := fmt.Fprint(s.config.Writer, "\r"+strings.Repeat(" ", s.lastPrintLen)+"\r")
	return err
}

func (s *Static) hideCursor() error {
	if !s.config.HideCursor {
		return nil
	}

	_, err := fmt.Fprint(s.config.Writer, "\r\033[?25l\r")
	return err
}

func (s *Static) unhideCursor() error {
	if !s.config.HideCursor {
		return nil
	}

	_, err := fmt.Fprint(s.config.Writer, "\r\033[?25h\r")
	return err
}
