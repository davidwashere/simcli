package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/davidwashere/simcli/internal/config"
	"github.com/davidwashere/simcli/internal/tasks"
)

const (
	Forever        = "forever"
	StdOutTaskType = "stdout"
	StdErrTaskType = "stderr"
	FileTaskType   = "file"
	HangTaskType   = "hang"
)

var (
	handlers = map[string]tasks.TaskHandler{}

	validTaskTypes = []string{
		StdOutTaskType,
		StdErrTaskType,
		FileTaskType,
		HangTaskType,
	}
)

func main() {
	c := config.Load()

	handlers[StdOutTaskType] = &tasks.StdOutTaskHandler{
		BasePath: c.BasePath,
	}
	handlers[StdErrTaskType] = &tasks.StdErrTaskHandler{
		BasePath: c.BasePath,
	}
	handlers[FileTaskType] = &tasks.FileTaskHandler{
		BasePath: c.BasePath,
	}
	handlers[HangTaskType] = &tasks.HangTaskHandler{}

	execute(c)
}

func execute(cfg *config.Config) {
	cmd, ok := matchCommand(cfg)

	// No command matches the arguements
	if !ok {
		if cfg.DefaultCommand == nil {
			log.Fatalf("ERROR: command not found for `%v` and no default command specified", cfg.Args)
		}
		handleCommand(cfg, cfg.DefaultCommand)
		return
	}

	handleCommand(cfg, cmd)
}

func matchCommand(cfg *config.Config) (*config.ConfigCommand, bool) {
	cmd, ok := cfg.CommandsM[cfg.Args]

	if ok {
		return cmd, ok
	}

	for _, cmd := range cfg.Commands {
		if cmd.Match == config.MatchContains {
			if strings.Contains(cfg.Args, cmd.Args) {
				return &cmd, true
			}
		}
	}

	return nil, false
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

		handler, ok := handlers[task.Type]

		if !ok {
			log.Fatalf("unrecognized task type '%v', valid types: %v", task.Type, validTaskTypes)
		}

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
