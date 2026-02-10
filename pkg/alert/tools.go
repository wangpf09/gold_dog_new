package alert

import (
	"github.com/wangpf09/golddog/pkg/metrics"
	"github.com/wangpf09/golddog/pkg/source"
)

func priceChange(window *metrics.RollingWindow[source.Derived]) []float64 {
	var p []float64
	for _, v := range window.Values() {
		p = append(p, v.PriceChange)
	}
	return p
}
