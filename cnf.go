package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	TelegramToken        string  `yaml:"telegramToken"`
	BotSummary           string  `yaml:"botSummary"`
	OllamaUrl            string  `yaml:"ollamaUrl"`
	Model                string  `yaml:"model"`
	HistoryRetainMinutes float32 `yaml:"historyRetainMinutes"`
}

var gConfig Config

func NewConfig() Config {
	var cnf Config
	cnf.TelegramToken = ""
	cnf.BotSummary = "Provide very brief, concise responses."
	cnf.OllamaUrl = "http://127.0.0.1:11434"
	cnf.Model = "llama3.2"
	cnf.TelegramToken = ""
	cnf.HistoryRetainMinutes = 8 * 60
	return cnf
}

var cnfConfigFileName = flag.String("config", "./tolby.yaml", "Config file name")
var cnfVerbose = flag.Bool("debug", false, "Debug verbose logs")

//var cnfGenConfig = flag.Bool("gen-config", false, "Generate default config.yaml to stdout")

func LoadConfigurationFile(filename string) Config {
	if verbose {
		log.Printf("Loading config file: %s", filename)
	}

	//1. Initialize config with default values
	config := NewConfig()

	//2. Load configuration files
	var loadedConfig Config
	configFile, err := os.Open(filename)

	if err != nil {
		log.Printf("Warning: '%s' config file not found, using default values!", filename)
		return config
		//log.Fatal(err.Error())
	}
	defer configFile.Close()

	jsonParser := yaml.NewDecoder(configFile)
	if jsonParser.Decode(&loadedConfig) != nil {
		log.Fatal(err.Error())
	}

	//3. Overwrite(Merge) default config with loaded config values
	config.TelegramToken = loadedConfig.TelegramToken
	if loadedConfig.Model != "" {
		config.Model = loadedConfig.Model
	}
	if loadedConfig.OllamaUrl != "" {
		config.OllamaUrl = loadedConfig.OllamaUrl
	}
	if loadedConfig.HistoryRetainMinutes > 0 {
		config.HistoryRetainMinutes = loadedConfig.HistoryRetainMinutes
	}

	return config
}
func initConfiguration() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options...]\n", os.Args[0])
		fmt.Fprint(os.Stderr, "\n")
		fmt.Fprint(os.Stderr, "Options:\n")
		fmt.Fprint(os.Stderr, "\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	gConfig = LoadConfigurationFile(*cnfConfigFileName)
}
