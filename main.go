package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	ConfigEnvKey      = "SIMCLI_CONFIG"
	DefaultConfigFile = "simcli.yaml"
)

type Config struct {
	Responses       []ConfigResponse `yaml:"responses"`
	Commands        []ConfigCommand  `yaml:"commands"`
	ResponsesM      map[string]*ConfigResponse
	Args            string
	DefaultResponse string `yaml:"defaultResponse"`
}

type ConfigResponse struct {
	Name       string
	Desc       string
	Input      string
	ReturnCode int `yaml:"rc"`
	Delay      int
}

type ConfigCommand struct {
	Args         string
	ResponseName string `yaml:"responseName"`
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

	config.ResponsesM = map[string]*ConfigResponse{}
	for i, r := range config.Responses {
		config.ResponsesM[r.Name] = &config.Responses[i]
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
	// fmt.Printf("%+v\n", config)

	doIt(config)
}

func doIt(config *Config) {
	for _, cmd := range config.Commands {
		if cmd.Args == config.Args {
			r, ok := config.ResponsesM[cmd.ResponseName]
			if !ok {
				break
			}

			printResponse(r)
			os.Exit(r.ReturnCode)
		}
	}

	printResponse(config.ResponsesM[config.DefaultResponse])
}

func printResponse(r *ConfigResponse) {
	file, err := os.Open(r.Input)
	check(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		fmt.Println(scanner.Text())
		if r.Delay > 0 {
			time.Sleep(time.Duration(r.Delay) * time.Millisecond)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
