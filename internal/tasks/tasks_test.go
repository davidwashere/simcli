package tasks

import (
	"testing"
)

func TestFileTaskHandle(t *testing.T) {

	fileTask := Task{
		Name:        "file-task",
		Type:        "file",
		Input:       "../../data/hello.txt",
		Delay:       0,
		InitDelay:   0,
		Repeat:      "",
		OutPath:     "test-output.txt",
		Permissions: 0755,
	}

	var handler TaskHandler

	handler = &FileTaskHandler{}
	handler.Handle(&fileTask)
}
