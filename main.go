package main

import (
	"context"

	"github.com/wangpf09/golddog/pkg/config"
	"github.com/wangpf09/golddog/pkg/logger"
	"github.com/wangpf09/golddog/pkg/monitor"
)

func main() {
	// Load configuration
	if err := config.LoadConfig("conf/config.yaml"); err != nil {
		panic("Failed to load config")
	}

	if err := logger.InitLogger(config.GetConfig().LoggerConfig); err != nil {
		panic(err)
	}

	ctx := context.Background()
	m, err := monitor.NewMonitor(config.GetConfig())
	if err != nil {
		panic(err)
	}

	if err := m.Run(ctx); err != nil {
		panic(err)
	}
}
