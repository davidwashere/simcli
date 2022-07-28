package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
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
	handlers = map[string]func(*ConfigTask){
		SysOutTaskType: handleSysOutErrTask,
		SysErrTaskType: handleSysOutErrTask,
		FileTaskType:   handleFileTask,
		HangTaskType:   handleHangTask,
	}
)

type Config struct {
	Tasks          []ConfigTask `yaml:"tasks"`
	TasksM         map[string]*ConfigTask
	Commands       []ConfigCommand `yaml:"commands"`
	CommandsM      map[string]*ConfigCommand
	Args           string
	DefaultCommand *ConfigCommand `yaml:"defaultCommand"`
}

type ConfigTask struct {
	Name      string
	Type      string
	Input     string
	Delay     int
	InitDelay int `yaml:"initdelay"`
	Repeat    string
	OutPath   string `yaml:"outPath"`
}

type ConfigCommand struct {
	Args       string
	Tasks      []string
	ReturnCode int `yaml:"rc"`
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
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

	config.TasksM = map[string]*ConfigTask{}
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

func main() {
	config := loadConfig()
	loadArgs(config)

	doIt(config)
}

func doIt(config *Config) {
	cmd, ok := config.CommandsM[config.Args]
	if !ok {
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
					handler(task)
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
			handler(task)
		}
	}

	os.Exit(cmd.ReturnCode)
}

func handleSysOutErrTask(t *ConfigTask) {
	file, err := os.Open(t.Input)
	check(err)
	defer file.Close()

	writer := os.Stdout
	if t.Type == SysErrTaskType {
		writer = os.Stderr
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if t.Delay > 0 {
			time.Sleep(time.Duration(t.Delay) * time.Millisecond)
		}
		fmt.Fprintln(writer, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func handleFileTask(t *ConfigTask) {
	iFile, err := os.Open(t.Input)
	check(err)
	defer iFile.Close()

	oFile, err := os.Create(t.OutPath)
	check(err)
	defer oFile.Close()

	_, err = io.Copy(oFile, iFile)
	check(err)
}

func handleHangTask(t *ConfigTask) {
	for {
		time.Sleep(time.Duration(1<<63 - 1))
	}
}
