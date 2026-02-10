package alert

import (
	"encoding/json"
	"fmt"
	"time"
)

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeJump       AlertType = "Jump"
	AlertTypeTrend      AlertType = "Trend"
	AlertTypeVolatility AlertType = "Volatility"
	AlertTypeHealth     AlertType = "Health"
)

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "Info"
	SeverityWarning  AlertSeverity = "Warning"
	SeverityCritical AlertSeverity = "Critical"
)

// AlertEvent represents a triggered alert with comprehensive information
type AlertEvent struct {
	Type      AlertType     `json:"type"`      // Alert type
	Severity  AlertSeverity `json:"severity"`  // Severity level
	Symbol    string        `json:"symbol"`    // Symbol that triggered
	Message   string        `json:"message"`   // Human-readable message
	Timestamp time.Time     `json:"timestamp"` // When triggered
}

// String returns a formatted string representation of the alert
func (a *AlertEvent) String() string {
	return fmt.Sprintf("[%s] [%s] %s Alert - %s: %s",
		a.Timestamp.Format("2006-01-02 15:04:05"),
		a.Severity,
		a.Type,
		a.Symbol,
		a.Message)
}

// ToJSON converts the alert to JSON bytes
func (a *AlertEvent) ToJSON() ([]byte, error) {
	return json.Marshal(a)
}

// ToFeishuCard converts the alert to Feishu card format
func (a *AlertEvent) ToFeishuCard() map[string]interface{} {
	// Determine color based on severity
	color := "grey"
	switch a.Severity {
	case SeverityCritical:
		color = "red"
	case SeverityWarning:
		color = "orange"
	case SeverityInfo:
		color = "blue"
	}

	// Determine emoji based on type
	emoji := "📊"
	switch a.Type {
	case AlertTypeJump:
		emoji = "🚨"
	case AlertTypeTrend:
		emoji = "📈"
	case AlertTypeHealth:
		emoji = "⚠️"
	case AlertTypeVolatility:
		emoji = "⚡"
	}

	// Build header
	title := fmt.Sprintf("%s %s Alert - %s", emoji, a.Type, a.Symbol)

	// Build fields
	fields := []map[string]interface{}{
		{
			"is_short": true,
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("**Severity**\n%s", a.Severity),
			},
		},
		{
			"is_short": true,
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("**Time**\n%s", a.Timestamp.Format("15:04:05")),
			},
		},
	}

	// Add message
	fields = append(fields, map[string]interface{}{
		"is_short": false,
		"text": map[string]interface{}{
			"tag":     "lark_md",
			"content": fmt.Sprintf("**Message**\n%s", a.Message),
		},
	})

	// Build card
	card := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"config": map[string]interface{}{
				"wide_screen_mode": true,
			},
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"tag":     "plain_text",
					"content": title,
				},
				"template": color,
			},
			"elements": []map[string]interface{}{
				{
					"tag":    "div",
					"fields": fields,
				},
				{
					"tag": "hr",
				},
				{
					"tag": "note",
					"elements": []map[string]interface{}{
						{
							"tag":     "plain_text",
							"content": fmt.Sprintf("Alert ID: %s-%d", a.Symbol, a.Timestamp.Unix()),
						},
					},
				},
			},
		},
	}

	return card
}
