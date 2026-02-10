package alert

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/wangpf09/golddog/pkg/logger"
	"github.com/wangpf09/golddog/pkg/metrics"
	"github.com/wangpf09/golddog/pkg/source"
)

// JumpDetector detects price jump (spike) events
type JumpDetector struct {
	threshold  float64       // Minimum price change to trigger alert (absolute or percent)
	cooldown   time.Duration // Minimum time between alerts for same symbol
	usePercent bool          // If true, threshold is percentage; if false, absolute
	mu         sync.RWMutex  // Protects symbolStates
}

// NewJumpDetector creates a new jump detector
func NewJumpDetector() *JumpDetector {
	return &JumpDetector{
		threshold:  1.5,
		cooldown:   time.Minute * 1,
		usePercent: true,
	}
}

// Evaluate evaluates a snapshot and returns an alert event if conditions are met
// Returns nil if no alert should be triggered
func (d *JumpDetector) Evaluate(window *metrics.RollingWindow[source.Derived]) *AlertEvent {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.zJump(window)
}

// JumpAlert z_jump = |price_change - mean(price_change)| / stddev(price_change)
func (d *JumpDetector) zJump(window *metrics.RollingWindow[source.Derived]) *AlertEvent {
	values := priceChange(window)
	std := metrics.StdDev(values)
	if std < 0.2 { // 防抖
		return nil
	}

	z := math.Abs(values[len(values)-1]-metrics.Mean(values)) / std

	logger.Debugf("z std: %.2f, lat: %.2f", std, z)

	if z >= 4.0 {
		return &AlertEvent{
			Type:      AlertTypeJump,
			Severity:  SeverityCritical,
			Message:   fmt.Sprintf("price jump detected: Δp=%.2f, z=%.2f", values[len(values)-1], z),
			Timestamp: time.Now(),
		}
	}
	return nil
}
