package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type StorageConfig struct {
	WhatsAppDB        string `yaml:"whatsapp_db"`
	SignalConfigDir   string `yaml:"signal_config_dir"`
	TempAttachmentDir string `yaml:"temp_attachment_dir"`
	SignalCLIPath     string `yaml:"signal_cli_path"`
}

type AccountsConfig struct {
	SignalNumber    string `yaml:"signal_number"`
	WhatsAppUserJID string `yaml:"whatsapp_user_jid"`
}

type Config struct {
	Storage    StorageConfig     `yaml:"storage"`
	Accounts   AccountsConfig    `yaml:"accounts"`
	GroupLinks map[string]string `yaml:"group_links"`
	Debug      bool              `yaml:"-"`
}

func DefaultConfig() *Config {
	return &Config{
		Storage: StorageConfig{
			WhatsAppDB:        "./data/whatsapp.db",
			SignalConfigDir:   "./data/signal",
			TempAttachmentDir: "./data/tmp",
			SignalCLIPath:     "signal-cli",
		},
		GroupLinks: make(map[string]string),
	}
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		// Return default config if file doesn't exist
		cfg := DefaultConfig()
		// Ensure directory exists
		if errDir := os.MkdirAll(filepath.Dir(path), 0700); errDir != nil {
			return nil, errDir
		}
		return cfg, nil
	} else if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := DefaultConfig()
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		return nil, err
	}

	if cfg.GroupLinks == nil {
		cfg.GroupLinks = make(map[string]string)
	}

	return cfg, nil
}

func SaveConfig(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	return encoder.Encode(cfg)
}
