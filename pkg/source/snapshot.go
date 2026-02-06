package source

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qos-max/qos-quote-api-go-sdk/qosapi"
)

// NormalizedSnapshot contains typed numeric fields converted from raw data
type NormalizedSnapshot struct {
	Symbol    string
	LastPrice float64
	Open      float64
	High      float64
	Low       float64
	Volume    float64
	Turnover  float64
	Timestamp time.Time
	Status    int // 0=normal, 1=suspended
}

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
