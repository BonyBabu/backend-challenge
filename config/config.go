package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/stretchr/testify/assert/yaml"
)

type Config struct {
	Db         string   `yaml:"db" validate:"required"`
	CouponBase []string `yaml:"couponBase" validate:"required"`
	CouponMin  int      `yaml:"couponMin" validate:"required"`
}

func GetConfig() Config {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Failed to Setup DB")
	}
	currentFileDir := filepath.Dir(filename)
	currentFileDir = filepath.ToSlash(currentFileDir)
	configFile := filepath.Join(currentFileDir, "../config.yaml")
	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}
	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Error unmarshalling YAML: %v", err)
	}
	if len(config.CouponBase) < config.CouponMin {
		log.Fatalf("Minimum %s couponBase file are required", config.CouponMin)
	}
	for _, file := range config.CouponBase {
		if _, err := os.Stat(file); err != nil {
			log.Fatalf("couponBase: %s doesn't exist", file)
		}
	}
	return config
}
