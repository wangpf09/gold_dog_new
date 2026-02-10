package metrics

import "math"

func Mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}

func StdDev(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	m := Mean(xs)
	var v float64
	for _, x := range xs {
		d := x - m
		v += d * d
	}
	return math.Sqrt(v / float64(len(xs)))
}
