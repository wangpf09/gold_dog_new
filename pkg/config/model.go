package config

import (
	"fmt"
	"time"
)

// Config represents the application configuration
type Config struct {
	LoggerConfig *LoggerConfig   `yaml:"logger"`
	QOSConfig    *QOSConfig      `yaml:"qos"`
	Alerts       *AlertConfig    `yaml:"alerts"`
	Notifier     *NotifierConfig `yaml:"notifier"`
}

// LoggerConfig 表示日志配置
type LoggerConfig struct {
	Filename   string `yaml:"filename"`
	Level      string `yaml:"level"` // debug info warn error panic fatal
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
	Console    bool   `yaml:"console"`
}

type QOSConfig struct {
	APIKey        string   `yaml:"api_key"`
	Symbols       []string `yaml:"symbols"`
	Heartbeat     int      `yaml:"heartbeat"`
	ChannelBuffer int      `yaml:"channel_buffer"`
}

// validate checks if the configuration is valid
func (c *QOSConfig) validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}

	if len(c.Symbols) == 0 {
		return fmt.Errorf("at least one symbol is required")
	}

	return nil
}

// GetHeartbeatDuration 获取心跳间隔时间
func (c *QOSConfig) GetHeartbeatDuration() time.Duration {
	return time.Duration(c.Heartbeat) * time.Second
}

// AlertConfig contains alert threshold settings
type AlertConfig struct {
	Jump   JumpConfig   `yaml:"jump"`
	Trend  TrendConfig  `yaml:"trend"`
	Health HealthConfig `yaml:"health"`
}

// JumpConfig defines configuration for Jump detector
type JumpConfig struct {
	Enabled    bool          `yaml:"enabled"`
	Threshold  float64       `yaml:"threshold"`
	UsePercent bool          `yaml:"use_percent"`
	Cooldown   time.Duration `yaml:"cooldown"`
}

// TrendConfig defines configuration for Trend detector
type TrendConfig struct {
	Enabled             bool          `yaml:"enabled"`
	EMAAlpha            float64       `yaml:"ema_alpha"`
	MinDuration         time.Duration `yaml:"min_duration"`
	Cooldown            time.Duration `yaml:"cooldown"`
	SuppressionDuration time.Duration `yaml:"suppression_duration"`
	MinOffsetThreshold  float64       `yaml:"min_offset_threshold"`
}

// HealthConfig defines configuration for Health detector
type HealthConfig struct {
	Enabled              bool          `yaml:"enabled"`
	MaxUnchangedPrice    int           `yaml:"max_unchanged_price"`
	MaxUnchangedVolume   int           `yaml:"max_unchanged_volume"`
	MaxUnchangedTurnover int           `yaml:"max_unchanged_turnover"`
	TimestampTolerance   time.Duration `yaml:"timestamp_tolerance"`
	CheckSuspended       bool          `yaml:"check_suspended"`
}

// NotifierConfig defines configuration for alert notifier
type NotifierConfig struct {
	Enabled        bool          `yaml:"enabled"`
	WebhookURL     string        `yaml:"webhook_url"`
	MaxRetries     int           `yaml:"max_retries"`
	InitialBackoff time.Duration `yaml:"initial_backoff"`
	Backoff        int           `yaml:"backoff"`
	Timeout        time.Duration `yaml:"timeout"`
	QueueSize      int           `yaml:"queue_size"`
	Workers        int           `yaml:"workers"`
}
