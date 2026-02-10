package source

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qos-max/qos-quote-api-go-sdk/qosapi"
)

// NormalizedSnapshot contains typed numeric fields converted from raw data
type NormalizedSnapshot struct {
	Symbol       string
	LastPrice    float64
	LastPriceCNY float64
	Open         float64
	High         float64
	Low          float64
	Volume       float64
	Turnover     float64
	Timestamp    time.Time
	Status       int // 0=normal, 1=suspended
}

const (
	// Currency conversion
	usdToCnyRate     = 6.92    // 美元转人民币汇率
	troyOunceToGrams = 31.1035 // 1盎司 = 31.1035克
)

// FromWSSnapshot converts a RawSnapshot (qosapi.WSSnapshot) to NormalizedSnapshot
func FromWSSnapshot(snapshot qosapi.WSSnapshot) (NormalizedSnapshot, error) {
	normalized := NormalizedSnapshot{
		Symbol:    snapshot.Code,
		Timestamp: time.Unix(snapshot.Timestamp, 0),
		Status:    snapshot.Suspended,
	}

	// Convert all string fields to float64
	var err error

	normalized.LastPrice, err = parseFloat(snapshot.LastPrice, "lp")
	if err != nil {
		return normalized, err
	}

	normalized.LastPriceCNY = (normalized.LastPrice * usdToCnyRate) / troyOunceToGrams

	normalized.Open, err = parseFloat(snapshot.Open, "o")
	if err != nil {
		return normalized, err
	}

	normalized.High, err = parseFloat(snapshot.High, "h")
	if err != nil {
		return normalized, err
	}

	normalized.Low, err = parseFloat(snapshot.Low, "l")
	if err != nil {
		return normalized, err
	}

	normalized.Volume, err = parseFloat(snapshot.Volume, "v")
	if err != nil {
		return normalized, err
	}

	normalized.Turnover, err = parseFloat(snapshot.Turnover, "t")
	if err != nil {
		return normalized, err
	}

	return normalized, nil
}

// parseFloat helper to convert string to float64 with error context
func parseFloat(s string, fieldName string) (float64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty value for field %s", fieldName)
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s: %w", fieldName, err)
	}

	return val, nil
}

type Derived struct {
	PriceChange     float64 // Δp
	PriceChangeRate float64
	VolumeDelta     float64 // Δv
}

func NewDerived(lastSnapshot, snapshot NormalizedSnapshot) Derived {
	priceChange := snapshot.LastPrice - lastSnapshot.LastPrice
	priceChangeRate := priceChange / lastSnapshot.LastPrice
	volumeDelta := snapshot.Volume - lastSnapshot.Volume
	return Derived{
		PriceChange:     priceChange,
		PriceChangeRate: priceChangeRate,
		VolumeDelta:     volumeDelta,
	}
}
