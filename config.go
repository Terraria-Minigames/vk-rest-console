package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		Port             int    `yaml:"Port"`
		TShockConfigPath string `yaml:"TShockConfigPath"`
		CommandPrefix    string `yaml:"CommandPrefix,omitempty"`
		RestAddr         string `yaml:"RestAddr,omitempty"` // If left empty, will default to "http://127.0.0.1:%d", where %d is filled from TShock config
		RemoveChatTags   bool   `yaml:"RemoveChatTags"`

		VK       ConfigVK       `yaml:"VK"`
		Messages ConfigMessages `yaml:"Messages"`
	}

	ConfigVK struct {
		ConfirmationToken string `yaml:"ConfirmationToken"`
		Secret            string `yaml:"Secret"`
		Keyboard          any    `yaml:"Keyboard,omitempty"`
		Token             string `yaml:"Token"`
	}

	ConfigMessages struct {
		NoCommandOutput   string `yaml:"NoCommandOutput"`
		RestRequestFailed string `yaml:"RestRequestFailed"`
	}

	TShockConfig struct {
		RestApiEnabled        bool
		RestApiPort           int
		CommandSpecifier      string
		ApplicationRestTokens map[string]struct {
			Username      string
			UserGroupName string
			VKId          int
		}
	}
)

func ValidateConfigValues(c Config) error {
	if c.TShockConfigPath == "" {
		return errors.New("TShockConfigPath is not set. Make sure --config-path argument is right and TShockConfigPath is set there")
	}
	if c.VK.Token == "" {
		return errors.New("VK group access token is not set in config")
	}
	return nil
}

func LoadConfig(path string, shouldCreate bool) (Config, error) {
	config := Config{
		Port:           80,
		RemoveChatTags: true,
		Messages: ConfigMessages{
			NoCommandOutput:   "Command didn't return anything.",
			RestRequestFailed: "REST Api malfunction, check Terraria Server logs",
		},
	}

	file, err := os.Open(path)
	if shouldCreate && errors.Is(err, fs.ErrNotExist) {
		file, err = os.Create(path)
		if err != nil {
			return config, err
		}
		err = yaml.NewEncoder(file).Encode(&config)
	}
	if err != nil {
		return config, err
	}
	defer file.Close()

	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		return config, err
	}

	return config, nil
}

func LoadTShockConfig(path string) (TShockConfig, error) {
	config := TShockConfig{}

	file, err := os.Open(path)
	if err != nil {
		return TShockConfig{}, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return config, err
	}

	// Reading post-4.5 TShock config format
	if err := json.Unmarshal(data, &struct{ Settings *TShockConfig }{&config}); err != nil {
		return config, err
	}

	// ApplicationRestTokens == nil means we didn't read anything, so it makes sense to fall back to pre-4.5 version format
	if config.ApplicationRestTokens == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			return config, err
		}
	}

	return config, nil
}
