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

	window := metrics.NewRollingWindow[source.NormalizedSnapshot](1000)

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
var ema = metrics.NewEMA(0.2)

func processSnapshot(snapshot source.NormalizedSnapshot, w *metrics.RollingWindow[source.NormalizedSnapshot]) {
	now := time.Now()

	if lastPush.IsZero() || now.Sub(lastPush) >= 12*time.Second {
		w.Push(snapshot)
		lastPush = now

		ema.Update(snapshot.LastPrice)
		if value, ok := ema.Value(); ok {
			slope := ema.Slope(12)
			logger.Debugf("Snapshot pushed, window size: %d,last price calc ema value: %.2f, slope value: %.2f", w.Size(), value, slope)
		}
	}

}
