package main

import (
	"context"
	"time"

	"github.com/wangpf09/golddog/pkg/config"
	"github.com/wangpf09/golddog/pkg/logger"
	"github.com/wangpf09/golddog/pkg/metrics"
	"github.com/wangpf09/golddog/pkg/source"
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
	sre, err := source.NewSnapshotSource(config.GetConfig().QOSConfig)
	if err != nil {
		panic(err)
	}
	if err = sre.Start(ctx); err != nil {
		panic(err)
	}

	window := metrics.NewRollingWindow[source.NormalizedSnapshot](100)

	for {
		select {
		case snapshot, ok := <-sre.Snapshots():
			if !ok {
				logger.Debugf("Snapshot channel closed")
				return
			}
			processSnapshot(snapshot, window)
		}
	}
}

var lastPush time.Time

func processSnapshot(snapshot source.NormalizedSnapshot, w *metrics.RollingWindow[source.NormalizedSnapshot]) {
	now := time.Now()

	if lastPush.IsZero() || now.Sub(lastPush) >= 12*time.Second {
		w.Push(snapshot)
		lastPush = now

		logger.Debugf("Snapshot pushed, window size: %d", w.Size())
	}

	logger.Debugf(
		"Snapshot received: lp=%.2f, time=%s",
		snapshot.LastPrice,
		snapshot.Timestamp.Format("2006-01-02 15:04:05"),
	)
}
