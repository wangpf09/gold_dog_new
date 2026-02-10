package alert

import (
	"fmt"
	"time"

	"github.com/wangpf09/golddog/pkg/logger"
	"github.com/wangpf09/golddog/pkg/metrics"
	"github.com/wangpf09/golddog/pkg/source"
)

type VolatilityDetector struct {
	consecutive int
}

// NewVolatilityDetector creates a new jump detector
func NewVolatilityDetector() *VolatilityDetector {
	return &VolatilityDetector{}
}

// Evaluate 10min的数据/60min的
func (v *VolatilityDetector) Evaluate(window *metrics.RollingWindow[source.Derived]) *AlertEvent {
	if window.Size() < 300 {
		return nil
	}

	values := priceChange(window)

	shortStd := metrics.StdDev(values[len(values)-50:])
	longStd := metrics.StdDev(values[len(values)-300:])

	if longStd < 0.2 {
		return nil
	}

	ratio := shortStd / longStd
	if ratio >= 2.5 {
		v.consecutive++
	} else {
		v.consecutive = 0
	}

	logger.Debugf("volatility short std: %.2f, long std: %.2f, ratio: %.2f", shortStd, longStd, ratio)

	if v.consecutive >= 3 {
		v.consecutive = 0
		return &AlertEvent{
			Type:      AlertTypeVolatility,
			Severity:  SeverityWarning,
			Message:   fmt.Sprintf("volatility increased: ratio=%.2f", ratio),
			Timestamp: time.Now(),
		}
	}
	return nil
}
