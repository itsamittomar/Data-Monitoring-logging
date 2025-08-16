package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Kafka struct {
		Brokers []string `yaml:"brokers"`
		Topic   string   `yaml:"topic"`
	} `yaml:"kafka"`
	
	Elasticsearch struct {
		URLs     []string `yaml:"urls"`
		Index    string   `yaml:"index"`
		Username string   `yaml:"username"`
		Password string   `yaml:"password"`
	} `yaml:"elasticsearch"`
	
	Metrics struct {
		Port         int    `yaml:"port"`
		Path         string `yaml:"path"`
		RetentionDays int   `yaml:"retention_days"`
	} `yaml:"metrics"`
	
	Alerting struct {
		RulesPath    string `yaml:"rules_path"`
		CheckInterval string `yaml:"check_interval"`
	} `yaml:"alerting"`
	
	Dashboard struct {
		Port int `yaml:"port"`
	} `yaml:"dashboard"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	
	return &config, nil
}