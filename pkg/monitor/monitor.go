package monitor

import (
	"context"
	"time"

	"github.com/wangpf09/golddog/pkg/alert"
	"github.com/wangpf09/golddog/pkg/config"
	"github.com/wangpf09/golddog/pkg/logger"
	"github.com/wangpf09/golddog/pkg/metrics"
	"github.com/wangpf09/golddog/pkg/notify"
	"github.com/wangpf09/golddog/pkg/source"
)

const (
	windowSize   = 7200
	pushInterval = 12 * time.Second
)

type Monitor struct {
	source *source.SnapshotSource

	jumpDetector       *alert.JumpDetector
	trendDetector      *alert.TrendDetector
	volatilityDetector *alert.VolatilityDetector

	notifier *notify.Notifier

	priceWindow       *metrics.RollingWindow[source.NormalizedSnapshot]
	priceChangeWindow *metrics.RollingWindow[source.Derived]

	lastPush     time.Time
	lastSnapshot source.NormalizedSnapshot
}

func NewMonitor(conf *config.Config) (*Monitor, error) {
	src, err := source.NewSnapshotSource(conf.QOSConfig)
	if err != nil {
		return nil, err
	}

	notifier, err := notify.NewNotifier(conf.Notifier)
	if err != nil {
		return nil, err
	}

	return &Monitor{
		source:             src,
		jumpDetector:       alert.NewJumpDetector(),
		trendDetector:      alert.NewTrendDetector(),
		volatilityDetector: alert.NewVolatilityDetector(),
		notifier:           notifier,
		priceWindow:        metrics.NewRollingWindow[source.NormalizedSnapshot](windowSize),
		priceChangeWindow:  metrics.NewRollingWindow[source.Derived](windowSize),
	}, nil
}

func (m *Monitor) Run(ctx context.Context) error {
	if err := m.source.Start(ctx); err != nil {
		return err
	}

	logger.Infof("📈 gold monitor started")
	_ = m.notifier.Send(&alert.AlertEvent{
		Type:      alert.AlertTypeHealth,
		Severity:  alert.SeverityInfo,
		Symbol:    "",
		Message:   "📈 gold monitor started",
		Timestamp: time.Now(),
	})
	for {
		select {
		case <-ctx.Done():
			return m.Close()

		case snap, ok := <-m.source.Snapshots():
			if !ok {
				logger.Warn("snapshot channel closed")
				return nil
			}
			m.handleSnapshot(snap)
		}
	}
}

func (m *Monitor) handleSnapshot(snap source.NormalizedSnapshot) {
	now := time.Now()

	if !m.lastPush.IsZero() && now.Sub(m.lastPush) < pushInterval {
		return
	}

	m.priceWindow.Push(snap)
	m.lastPush = now

	if m.priceWindow.Size() > 1 {
		m.priceChangeWindow.Push(
			source.NewDerived(m.lastSnapshot, snap),
		)
	}

	if m.priceChangeWindow.Size() > 2 {
		if !m.evaluate(snap) {
			logger.Debugf("current price: %.2f 元/克", snap.LastPriceCNY)
		}
	}

	m.lastSnapshot = snap
}

func (m *Monitor) evaluate(snap source.NormalizedSnapshot) bool {
	hasAlert := false

	if e := m.jumpDetector.Evaluate(m.priceChangeWindow); e != nil {
		m.dispatch(e)
		hasAlert = true
	}

	if e := m.trendDetector.Evaluate(snap.LastPrice); e != nil {
		m.dispatch(e)
		hasAlert = true
	}

	if e := m.volatilityDetector.Evaluate(m.priceChangeWindow); e != nil {
		m.dispatch(e)
		hasAlert = true
	}

	return hasAlert
}

func (m *Monitor) dispatch(e *alert.AlertEvent) {
	logger.Infof("🚨 ALERT: %s", e.String())

	if err := m.notifier.Send(e); err != nil {
		logger.Warnf("failed to send alert: %v", err)
	}
}

func (m *Monitor) Close() error {
	logger.Info("monitor shutting down")

	if m.notifier != nil {
		m.notifier.Close()
	}
	return nil
}
