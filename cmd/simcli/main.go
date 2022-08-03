package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/davidwashere/simcli/internal/tasks"
)

const (
	ConfigEnvKey      = "SIMCLI_CONFIG"
	DefaultConfigFile = "simcli.yaml"
	Forever           = "forever"
	SysErrTaskType    = "syserr"
	SysOutTaskType    = "sysout"
	FileTaskType      = "file"
	HangTaskType      = "hang"
)

var (
	handlers = map[string]tasks.TaskHandler{
		SysOutTaskType: &tasks.SysOutTaskHandler{},
		SysErrTaskType: &tasks.SysErrTaskHandler{},
		FileTaskType:   &tasks.FileTaskHandler{},
		HangTaskType:   &tasks.HangTaskHandler{},
	}
)

type Config struct {
	Tasks          []tasks.Task `yaml:"tasks"`
	TasksM         map[string]*tasks.Task
	Commands       []ConfigCommand `yaml:"commands"`
	CommandsM      map[string]*ConfigCommand
	Args           string
	DefaultCommand *ConfigCommand `yaml:"defaultCommand"`
}

type ConfigCommand struct {
	Args       string
	Tasks      []string
	ReturnCode int `yaml:"rc"`
}

func main() {
	config := loadConfig()
	loadArgs(config)
	doIt(config)
}

func loadConfig() *Config {
	pathFromEnv := os.Getenv(ConfigEnvKey)

	var err error
	var fileB []byte

	configFilePath := DefaultConfigFile
	if len(pathFromEnv) > 0 {
		configFilePath = pathFromEnv
	}

	fileB, err = os.ReadFile(configFilePath)
	if err != nil {
		log.Fatalf("failed to open config: %v", err)
	}

	config := Config{}
	err = yaml.Unmarshal(fileB, &config)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	config.TasksM = map[string]*tasks.Task{}
	for i, r := range config.Tasks {
		config.TasksM[r.Name] = &config.Tasks[i]
	}

	config.CommandsM = map[string]*ConfigCommand{}
	for i, c := range config.Commands {
		config.CommandsM[c.Args] = &config.Commands[i]
	}

	return &config
}

func loadArgs(config *Config) {
	args := strings.Join(os.Args[1:], " ")
	config.Args = args
}

func doIt(config *Config) {
	cmd, ok := config.CommandsM[config.Args]
	if !ok {
		if config.DefaultCommand == nil {
			log.Fatalf("ERROR: command not found for `%v` and no default command specified", config.Args)
		}
		handleCommand(config, config.DefaultCommand)
		return
	}

	handleCommand(config, cmd)
}

func handleCommand(config *Config, cmd *ConfigCommand) {
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
