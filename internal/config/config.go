package config

import (
	"log"
	"sync"

	"github.com/Falokut/intern_test_task/internal/repository"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel string `yaml:"log_level" env:"LOG_LEVEL"`
	Listen   struct {
		Host string `yaml:"host" env:"HOST"`
		Port string `yaml:"port" env:"PORT"`
	} `yaml:"listen"`

	DBConfig repository.DBConfig `yaml:"db_config"`
}

var instance *Config
var once sync.Once

const configsPath = "configs/"

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}

		if err := cleanenv.ReadConfig(configsPath+"config.yml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			log.Fatal(help, " ", err)
		}
	})

	return instance
}
