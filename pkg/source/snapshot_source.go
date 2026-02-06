package source

import (
	"context"
	"fmt"
	"sync"

	"github.com/qos-max/qos-quote-api-go-sdk/qosapi"

	"github.com/wangpf09/golddog/pkg/config"
	"github.com/wangpf09/golddog/pkg/logger"
)

// SnapshotSource manages WebSocket connection to qosapi and provides snapshot data
type SnapshotSource struct {
	cfg *config.QOSConfig

	client *qosapi.WSClient

	snapshots chan NormalizedSnapshot
	mu        sync.RWMutex
	started   bool
}

// NewSnapshotSource creates a new SnapshotSource instance
func NewSnapshotSource(cfg *config.QOSConfig) (*SnapshotSource, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	if len(cfg.Symbols) == 0 {
		return nil, fmt.Errorf("at least one symbol is required")
	}

	bufferSize := cfg.ChannelBuffer
	if bufferSize <= 0 {
		bufferSize = 100
	}

	return &SnapshotSource{
		cfg:       cfg,
		snapshots: make(chan NormalizedSnapshot, bufferSize),
	}, nil
}

// Start initializes the WebSocket client, connects, and starts receiving snapshots
func (s *SnapshotSource) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return fmt.Errorf("source already started")
	}
	s.started = true
	s.mu.Unlock()

	// Initialize qosapi WebSocket client
	s.client = qosapi.NewWSClient(s.cfg.APIKey)

	// Connect to WebSocket
	logger.Debugf("正在连接到 WebSocket 服务器...")
	logger.Debugf("Symbols: %v", s.cfg.Symbols)

	if err := s.client.Connect(); err != nil {
		s.mu.Lock()
		s.started = false
		s.mu.Unlock()

		// Provide detailed error information
		logger.Debugf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		logger.Debugf("❌ WebSocket 连接失败: %v", err)
		logger.Debugf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		return fmt.Errorf("failed to connect: %w", err)
	}

	logger.Debugf("WebSocket connected")

	// Start heartbeat to keep connection alive (every 30 seconds)
	logger.Debugf("Starting heartbeat...")
	s.client.StartHeartbeat(s.cfg.GetHeartbeatDuration())

	// Subscribe to symbols for snapshot data
	if err := s.subscribeSymbols(); err != nil {
		s.client.Close()
		s.mu.Lock()
		s.started = false
		s.mu.Unlock()
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Handle context cancellation
	go func() {
		<-ctx.Done()
		logger.Errorf("Context cancelled, closing connection...")
		s.Close()
	}()

	logger.Debugf("SnapshotSource started, monitoring %d symbols", len(s.cfg.Symbols))
	return nil
}

// subscribeSymbols subscribes to all configured symbols for snapshot data
func (s *SnapshotSource) subscribeSymbols() error {
	if len(s.cfg.Symbols) == 0 {
		return fmt.Errorf("no symbols to subscribe")
	}

	logger.Debugf("Subscribing to symbols: %v", s.cfg.Symbols)

	// Subscribe to snapshot data for all symbols with callback
	// The SDK supports batch subscription
	if err := s.client.SubscribeSnapshot(s.cfg.Symbols, s.handleSnapshot); err != nil {
		return fmt.Errorf("failed to subscribe to snapshots: %w", err)
	}

	logger.Debugf("Successfully subscribed to %d symbols", len(s.cfg.Symbols))
	return nil
}

// handleSnapshot processes incoming snapshot data (non-blocking)
func (s *SnapshotSource) handleSnapshot(wsSnapshot qosapi.WSSnapshot) {
	// Convert qosapi.WSSnapshot to internal RawSnapshot format

	// Convert to NormalizedSnapshot
	normalized, err := FromWSSnapshot(wsSnapshot)
	if err != nil {
		logger.Errorf("Failed to normalize snapshot for %s: %v", wsSnapshot.Code, err)
		return
	}

	// Non-blocking send to channel
	select {
	case s.snapshots <- normalized:
		// Successfully sent
	default:
		// Channel full, drop oldest or log warning
		logger.Warnf("Warning: snapshot channel full, dropping snapshot for %s", normalized.Symbol)
	}
}

// Snapshots returns a read-only channel for receiving normalized snapshots
func (s *SnapshotSource) Snapshots() <-chan NormalizedSnapshot {
	return s.snapshots
}

// Close gracefully shuts down the WebSocket connection and closes channels
func (s *SnapshotSource) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return nil
	}

	logger.Debugf("Closing SnapshotSource...")

	if s.client != nil {
		s.client.Close()
	}

	// Close the snapshot channel
	close(s.snapshots)
	s.started = false

	logger.Debugf("SnapshotSource closed")
	return nil
}
