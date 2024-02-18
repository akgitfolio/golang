package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/asaskevich/govalidator"
	"gopkg.in/yaml.v2"
)

type Config struct {
	APIKey       string `valid:"required,alphanum,length(32)"`
	DatabaseHost string `valid:"required,hostname"`
	DatabasePort int    `valid:"required,port"`
	DebugMode    bool   `valid:"optional"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing configuration file: %w", err)
	}

	_, err = govalidator.ValidateStruct(&config)
	if err != nil {
		return nil, fmt.Errorf("configuration file validation failed: %w", err)
	}

	return &config, nil
}

func main() {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	config, err := LoadConfig("config.yaml")
	if err != nil {
		panic(err)
	}

	switch environment {
	case "development":
	case "staging":
	case "production":
	}

	fmt.Printf("API Key: %s\n", config.APIKey)
	fmt.Printf("Database Host: %s\n", config.DatabaseHost)
	fmt.Printf("Database Port: %d\n", config.DatabasePort)
	fmt.Printf("Debug Mode: %t\n", config.DebugMode)
}
