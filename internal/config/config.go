package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type AIChannel string
type SourceType string

const (
	AIChannelCompatibleOpenAI AIChannel = "compatible-openai"
)
const (
	SourceTypeFile     SourceType = "file"
	SourceTypeJournald SourceType = "journald"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server" validate:"required"`
	AI       AIConfig       `mapstructure:"ai" validate:"required"`
	Database DatabaseConfig `mapstructure:"database" validate:"required"`
	Query    QueryConfig    `mapstructure:"query" validate:"required"`
	Logs     LogsConfig     `mapstructure:"logs"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port" validate:"required,gt=0,lte=65535"`
	Mode string `mapstructure:"mode" validate:"required,oneof=debug release test"`
}

type AIConfig struct {
	APIKey      string        `mapstructure:"api_key" validate:"required"`
	Channel     AIChannel     `mapstructure:"channel" validate:"required,oneof=compatible-openai"`
	BaseURL     string        `mapstructure:"base_url" validate:"omitempty,url"`
	Model       string        `mapstructure:"model" validate:"required"`
	Prompt      string        `mapstructure:"prompt" validate:"required"`
	Timeout     time.Duration `mapstructure:"timeout" validate:"required,gt=0"`
	Temperature float64       `mapstructure:"temperature" validate:"gte=0,lte=2"`
	QueueSize   int           `mapstructure:"queue_size" validate:"gte=0"`
	WorkerCount int           `mapstructure:"worker_count" validate:"gte=1"`
}

type QueryConfig struct {
	RangeTime time.Duration `mapstructure:"range_time" validate:"required,gt=0"`
	Step      time.Duration `mapstructure:"step" validate:"required,gt=0"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn" validate:"required"`
}

type LogsConfig struct {
	MaxLines int            `mapstructure:"max_lines" validate:"gte=0"`
	Sources  []SourceConfig `mapstructure:"sources" validate:"dive"`
}

type SourceConfig struct {
	LabelKey   string `mapstructure:"label_key" validate:"required"`
	LabelValue string `mapstructure:"label_value" validate:"required"`
	FuzzyMatch bool   `mapstructure:"fuzzy_match"`
	Type       string `mapstructure:"type" validate:"required,oneof=file journald"`
	Path       string `mapstructure:"path" validate:"required_if=type file"`
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

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置文件校验失败: %w", err)
	}

	return &config, nil
}

func (c *Config) Validate() error {
	validate := validator.New()

	validate.RegisterStructValidation(validateAIConfig, AIConfig{})

	return validate.Struct(c)
}
func validateAIConfig(sl validator.StructLevel) {
	ai := sl.Current().Interface().(AIConfig)

	if strings.Contains(string(ai.Channel), "compatible") && strings.TrimSpace(ai.BaseURL) == "" {
		sl.ReportError(ai.BaseURL, "BaseURL", "base_url", "required_for_compatible_channel", "")
	}
}
