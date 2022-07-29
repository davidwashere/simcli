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
		if config.DefaultCommand == nil {
			log.Fatalf("ERROR: command not found for `%v` and no default command specified", config.Args)
		}
		handleCommand(config, config.DefaultCommand)
		return
	}

	handleCommand(config, cmd)
}

func handleCommand(config *Config, cmd *ConfigCommand) {
	log.Printf("handleCommand config %v cmd %v", config, cmd)
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

	batch := 1
	if t.Delay > 0 && t.Delay <= 15 {
		batch = 16 - t.Delay
	}

	batchBuf := make([]string, batch)

	cnt := 0
	for scanner.Scan() {
		if t.Delay == 0 {
			fmt.Fprintln(writer, scanner.Text())
			continue
		}

		// A batch size of 1 means no need for buffer to meet SLA, but is a delay
		if batch == 1 {
			fmt.Fprintln(writer, scanner.Text())
			time.Sleep(time.Duration(t.Delay) * time.Millisecond)
			continue
		}

		// batch size must be > 0, which means delay < 15ms
		batchBuf[cnt] = scanner.Text()
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

/*
Regarding batch sizes in HandleSysErrOutTask

time.Sleep on windows a was could not sleep less than 16 ms, see data for 3844 line file,
the real time was the actual time it was taking vs- what would be expected.

As a result add a line 'batch' for any delays < 16ms - the math isn't exact for batch size
but the #'s are much closer to expected when batching

  delay | real | expected (best case)
  --- | --- | ---
  1  | 0m59.638s | 3.844
  5  | 0m59.677s | 19.220
  10 | 0m59.699s | 38.440
  11 | 0m59.730s | 42.284
  12 | 0m59.711s | 46.128
  13 | 0m59.661s | 49.972
  14 | 0m59.743s | 53.816
  15 | 1m8.219s  | 57.66
  20 | 1m59.426s | 76.88 (1m16.880s)

*/
