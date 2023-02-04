package tasks

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Task struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Input       string `yaml:"input"`
	Delay       int    `yaml:"delay"`
	InitDelay   int    `yaml:"initdelay"`
	Repeat      string `yaml:"repeat"`
	OutPath     string `yaml:"outPath"`
	Permissions uint32 `yaml:"perms"`
}

type TaskHandler interface {
	Handle(t *Task) error
}

type FileTaskHandler struct {
	BasePath string
}
type HangTaskHandler struct{}

type StdOutTaskHandler struct {
	BasePath string
}

type StdErrTaskHandler struct {
	BasePath string
}

func (f *FileTaskHandler) Handle(t *Task) error {
	path := t.Input
	if !filepath.IsAbs(path) {
		path = filepath.Join(f.BasePath, t.Input)
	}

	iFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer iFile.Close()

	oFile, err := os.Create(t.OutPath)
	if err != nil {
		return err
	}
	defer oFile.Close()

	if t.Permissions != 0 {
		oFile.Chmod(fs.FileMode(t.Permissions))
	} else {
		oFile.Chmod(0644)
	}

	_, err = io.Copy(oFile, iFile)
	return err
}

func (h *HangTaskHandler) Handle(t *Task) error {
	for {
		time.Sleep(time.Duration(1<<63 - 1))
	}
}

func (h *StdOutTaskHandler) Handle(t *Task) error {
	return printWriter(h.BasePath, t, os.Stdout)
}

func (h *StdErrTaskHandler) Handle(t *Task) error {
	return printWriter(h.BasePath, t, os.Stderr)
}

func printWriter(basePath string, t *Task, writer io.Writer) error {
	path := t.Input
	if !filepath.IsAbs(path) {
		path = filepath.Join(basePath, t.Input)
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	batch := 1
	if t.Delay > 0 && t.Delay <= 15 {
		batch = 16 - t.Delay
	}

	batchBuf := make([]string, batch)

	cnt := 0
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "\n") {
			line = strings.ReplaceAll(line, "\r", "")
		} else {
			line = strings.ReplaceAll(line, "\r", "\n")
		}

		if t.Delay == 0 {
			fmt.Fprintln(writer, line)
			continue
		}

		// A batch size of 1 means no need for buffer to meet SLA, but is a delay
		if batch == 1 {
			fmt.Fprintln(writer, line)
			time.Sleep(time.Duration(t.Delay) * time.Millisecond)
			continue
		}

		// batch size must be > 0, which means delay < 15ms
		batchBuf[cnt] = line
		cnt++

		if cnt == batch {
			cnt = 0

			for _, item := range batchBuf {
				fmt.Fprintln(writer, item)
			}

			time.Sleep(time.Duration(16) * time.Millisecond)
		}
	}

	for i := 0; i < cnt; i++ {
		fmt.Fprintln(writer, batchBuf[i])
	}

	return scanner.Err()
}
