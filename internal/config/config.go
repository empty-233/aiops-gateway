package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	AI       AIConfig       `mapstructure:"ai"`
	Database DatabaseConfig `mapstructure:"database"`
	Query    QueryConfig    `mapstructure:"query"`
	Logs     LogsConfig     `mapstructure:"logs"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type AIConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
	Model   string `mapstructure:"model"`
	Timeout time.Duration `mapstructure:"timeout"`
	Temperature float64 `mapstructure:"temperature"`
}

type QueryConfig struct {
	RangeTime time.Duration `mapstructure:"range_time"`
	Step      time.Duration `mapstructure:"step"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

type LogsConfig struct {
	MaxLines int            `mapstructure:"max_lines"`
	Source   []SourceConfig `mapstructure:"sources"`
}

type SourceConfig struct {
	LabelKey   string `mapstructure:"label_key"`
	LabelValue string `mapstructure:"label_value"`
	FuzzyMatch bool   `mapstructure:"fuzzy_match"`
	Type       string `mapstructure:"type"`
	Path       string `mapstructure:"path"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	v.AutomaticEnv()
	v.SetEnvPrefix("AIOPS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("配置文件读取失败: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("配置文件解析失败: %w", err)
	}

	return &config, nil
}
