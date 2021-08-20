package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                int
	RestUrl             string
	TShockConfig        string
	VKUserTokens        map[int]string `json:"-"`
	VKConfirmationToken string
	VKSecret            string
	VKKeyboard          string
	VKToken             string
}

type TShockConfig struct {
	Settings struct {
		ApplicationRestTokens map[string](map[string]int)
	}
}

func MakeDefaultConfig() Config {
	config := Config{
		Port:                80,
		RestUrl:             "http://localhost:7878",
		TShockConfig:        "",
		VKUserTokens:        make(map[int]string),
		VKConfirmationToken: "",
		VKSecret:            "",
		VKKeyboard:          "",
		VKToken:             "",
	}

	return config
}

func LoadConfig() Config {
	config := MakeDefaultConfig()
	file, err := os.Open("config.json")

	if err != nil {
		jsondata, _ := json.MarshalIndent(config, "", "\t")
		ioutil.WriteFile("config.json", jsondata, 0644)
		fmt.Println("Config not found. Created default.")
	}

	defer file.Close()

	jsondata, _ := ioutil.ReadAll(file)
	json.Unmarshal(jsondata, &config)

	if config.TShockConfig == "" {
		fmt.Println("TShockConfig is not set")
		os.Exit(2)
	}

	LoadTShockTokens(config.TShockConfig, &config)

	// Make sure RestUrl does not end with /
	config.RestUrl = strings.TrimSuffix(config.RestUrl, "/")

	fmt.Println("Loaded config.")
	return config
}

func LoadTShockTokens(path string, config *Config) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
	defer file.Close()

	data, _ := ioutil.ReadAll(file)
	var tshockConfig TShockConfig
	json.Unmarshal(data, &tshockConfig)

	for k, v := range tshockConfig.Settings.ApplicationRestTokens {
		userId := v["VKId"]

		if userId == 0 {
			continue
		}

		fmt.Println("Loaded tshock token for " + strconv.Itoa(userId))
		config.VKUserTokens[userId] = k
	}
}
