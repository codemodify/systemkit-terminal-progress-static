package main

import (
	"time"

	static "github.com/codemodify/systemkit-terminal-progress-static"
)

func main() {
	successSpinner := static.NewStatic("Running operation 1")
	successSpinner.Run()
	time.Sleep(5 * time.Second)
	successSpinner.Success()

	failSpinner := static.NewStatic("Running operation 2")
	failSpinner.Run()
	time.Sleep(5 * time.Second)
	failSpinner.Fail()
}
