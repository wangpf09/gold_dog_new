package alert

import (
	"fmt"
	"math"
	"time"

	"github.com/wangpf09/golddog/pkg/logger"
	"github.com/wangpf09/golddog/pkg/metrics"
)

type TrendDetector struct {
	emaFast     *metrics.EMA
	emaSlow     *metrics.EMA
	consecutive int
}

func NewTrendDetector() *TrendDetector {
	return &TrendDetector{
		emaFast: metrics.NewEMA(0.2),
		emaSlow: metrics.NewEMA(0.05),
	}
}

func (t *TrendDetector) Evaluate(price float64) *AlertEvent {
	t.emaFast.Update(price)
	t.emaSlow.Update(price)
	var fast, slow float64

	if f, ok := t.emaFast.Value(); ok {
		fast = f
	}
	if s, ok := t.emaSlow.Value(); ok {
		slow = s
	}
	diff := fast - slow

	slope := t.emaFast.Slope(12)

	sameDirection := (diff > 0 && slope > 0) || (diff < 0 && slope < 0)

	if math.Abs(diff) >= 3.0 && math.Abs(slope) >= 0.0025 && sameDirection {
		t.consecutive++
	} else {
		t.consecutive = 0
	}
	
	logger.Debugf("trend ema fast: %.2f, slow: %.2f, slope: %.2f", fast, slow, slope)

	if t.consecutive >= 5 { // ≈1分钟
		t.consecutive = 0

		dir := "up"
		if slope < 0 {
			dir = "down"
		}

		return &AlertEvent{
			Type:      AlertTypeTrend,
			Severity:  SeverityInfo,
			Message:   fmt.Sprintf("trend %s detected, slope=%.4f USD/s, ema_diff=%.2f", dir, slope, diff),
			Timestamp: time.Now(),
		}
	}
	return nil
}
