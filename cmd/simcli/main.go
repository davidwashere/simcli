package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/davidwashere/simcli/internal/config"
	"github.com/davidwashere/simcli/internal/tasks"
)

const (
	Forever        = "forever"
	SysErrTaskType = "syserr"
	SysOutTaskType = "sysout"
	FileTaskType   = "file"
	HangTaskType   = "hang"
)

var (
	handlers = map[string]tasks.TaskHandler{
		SysOutTaskType: &tasks.SysOutTaskHandler{},
		SysErrTaskType: &tasks.SysErrTaskHandler{},
		FileTaskType:   &tasks.FileTaskHandler{},
		HangTaskType:   &tasks.HangTaskHandler{},
	}
)

func main() {
	c := config.Load()

	execute(c)
}

func execute(config *config.Config) {
	cmd, ok := config.CommandsM[config.Args]

	// No command matches the arguements
	if !ok {
		if config.DefaultCommand == nil {
			log.Fatalf("ERROR: command not found for `%v` and no default command specified", config.Args)
		}
		handleCommand(config, config.DefaultCommand)
		return
	}

	handleCommand(config, cmd)
}

func handleCommand(config *config.Config, cmd *config.ConfigCommand) {
	for _, taskName := range cmd.Tasks {
		task, ok := config.TasksM[taskName]
		if !ok {
			log.Fatalf("task %v not found", taskName)
		}

		if task.InitDelay > 0 {
			time.Sleep(time.Duration(task.InitDelay) * time.Millisecond)
		}

		handler := handlers[task.Type]

		repeats := 1
		if task.Repeat != "" {
			if task.Repeat == Forever {
				for {
					if err := handler.Handle(task); err != nil {
						log.Fatal(err)
					}
				}
			} else {
				var err error
				repeats, err = strconv.Atoi(task.Repeat)
				if err != nil {
					repeats = 1
				}
			}
		}

		for i := 0; i < repeats; i++ {
			if err := handler.Handle(task); err != nil {
				log.Fatal(err)
			}
		}
	}

	os.Exit(cmd.ReturnCode)
}
