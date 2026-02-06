package metrics

// EMA represents an Exponential Moving Average calculator
type EMA struct {
	alpha       float64
	value       float64
	preValue    float64
	initialized bool
}

// NewEMA creates a new EMA calculator with the specified alpha (smoothing factor)
// Alpha should be between 0 and 1. Higher alpha = more weight on recent values.
// Common values: 0.1, 0.2, or calculate from period: alpha = 2/(period+1)
func NewEMA(alpha float64) *EMA {
	if alpha <= 0 || alpha > 1 {
		alpha = 0.2 // Default
	}
	return &EMA{
		alpha:       alpha,
		initialized: false,
	}
}

// NewEMAFromPeriod creates an EMA with alpha calculated from a period
// For example, period=10 gives alpha = 2/(10+1) ≈ 0.1818
func NewEMAFromPeriod(period int) *EMA {
	if period <= 0 {
		period = 10
	}
	alpha := 2.0 / float64(period+1)
	return NewEMA(alpha)
}

// Update updates the EMA with a new value
func (e *EMA) Update(value float64) {
	if !e.initialized {
		e.value = value
		e.initialized = true
	} else {
		e.value = e.alpha*value + (1-e.alpha)*e.value
	}
}

// Value returns the current EMA value
func (e *EMA) Value() (float64, bool) {
	return e.value, e.initialized
}

// Reset resets the EMA to uninitialized state
func (e *EMA) Reset() {
	e.initialized = false
	e.value = 0
}

// IsInitialized returns true if the EMA has received at least one value
func (e *EMA) IsInitialized() bool {
	return e.initialized
}

// Alpha returns the smoothing factor
func (e *EMA) Alpha() float64 {
	return e.alpha
}
