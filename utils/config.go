package utils

import (
	"Qpan/models"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

func ConfigInit() models.Config {
	file, err := os.Open("config.yaml")
	if err != nil {
		fmt.Println("Error opening config file:", err)
		return models.Config{}
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)

	var config models.Config

	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error decoding config file:", err)
		return models.Config{}
	}
	models.ServerHost = config.Server.Host
	return config
}
