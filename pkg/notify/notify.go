package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wangpf09/golddog/pkg/alert"
	"github.com/wangpf09/golddog/pkg/config"
	"github.com/wangpf09/golddog/pkg/logger"
)

// Notifier 负责告警分发
type Notifier struct {
	cfg    *config.NotifierConfig
	client *http.Client
	queue  chan *alert.AlertEvent
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	closed atomic.Bool // 原子操作标记是否已关闭，防止向已关闭的 channel 发送数据
}

//type Config struct {
//	WebhookURL string
//	Workers    int
//	QueueSize  int
//
//	MaxRetries int
//	Backoff    time.Duration
//	Backoff time.Duration
//	Timeout    time.Duration
//}

func NewNotifier(cfg *config.NotifierConfig) (*Notifier, error) {
	if cfg.WebhookURL == "" {
		return nil, errors.New("webhook url required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	n := &Notifier{
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.Timeout},
		queue:  make(chan *alert.AlertEvent, cfg.QueueSize),
		ctx:    ctx,
		cancel: cancel,
	}

	n.startWorkers()
	return n, nil
}

func (n *Notifier) startWorkers() {
	for i := 0; i < n.cfg.Workers; i++ {
		n.wg.Add(1)
		go n.worker(i)
	}
}

// Send 非阻塞发送告警
func (n *Notifier) Send(a *alert.AlertEvent) error {
	// 检查是否已关闭，避免 panic
	if n.closed.Load() {
		return errors.New("notifier is closed")
	}

	select {
	case n.queue <- a:
		return nil
	default:
		return errors.New("alert queue full")
	}
}

func (n *Notifier) worker(id int) {
	defer n.wg.Done()
	logger.Debugf("[notifier] worker-%d started", id)

	// 使用 range 循环，这样 channel 关闭时会自动退出，
	// 并且会处理完 channel 中剩余的数据（优雅退出）
	for a := range n.queue {
		if err := n.handleAlert(a); err != nil {
			logger.Errorf("[notifier] worker-%d failed to send %s: %v", id, a.Symbol, err)
		}
	}
}

func (n *Notifier) handleAlert(a *alert.AlertEvent) error {
	// 优化1：只序列化一次
	payload, err := json.Marshal(a.ToFeishuCard())
	if err != nil {
		return err
	}

	var lastErr error
	// 循环次数 = 1次首次尝试 + MaxRetries次重试
	for i := 0; i <= n.cfg.MaxRetries; i++ {
		// 检查上下文是否已取消（快速退出）
		if n.ctx.Err() != nil {
			return n.ctx.Err()
		}

		if err := n.doRequest(payload); err == nil {
			return nil
		} else {
			lastErr = err
		}

		// 如果不是最后一次尝试，则等待
		if i < n.cfg.MaxRetries {
			wait := n.calcBackoff(i + 1)
			logger.Warnf("[notifier] retry %d/%d in %v (%s): %v", i+1, n.cfg.MaxRetries, wait, a.Symbol, lastErr)

			select {
			case <-time.After(wait):
			case <-n.ctx.Done(): // 支持重试等待期间被取消
				return n.ctx.Err()
			}
		}
	}
	return lastErr
}

func (n *Notifier) doRequest(body []byte) error {
	// 优化2：请求绑定 Context，以便 Close() 时能取消正在进行的 HTTP 请求
	req, err := http.NewRequestWithContext(n.ctx, "POST", n.cfg.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New(resp.Status)
	}
	return nil
}

// calcBackoff 计算带抖动的指数退避
func (n *Notifier) calcBackoff(attempt int) time.Duration {
	// 2^(attempt-1) * base
	factor := math.Pow(2, float64(attempt-1))
	backoff := float64(n.cfg.Backoff) * factor

	if backoff > float64(n.cfg.Backoff) {
		backoff = float64(n.cfg.Backoff)
	}

	// 优化3：简化抖动逻辑，+/- 10%
	jitter := (rand.Float64()*0.2 - 0.1) * backoff
	return time.Duration(backoff + jitter)
}

// Close 优雅关闭
func (n *Notifier) Close() {
	if !n.closed.CompareAndSwap(false, true) {
		return // 已经关闭
	}

	// 1. 关闭 channel，让 worker 处理完剩余数据后退出 range 循环
	close(n.queue)

	// 2. 等待所有 worker 处理完积压数据
	n.wg.Wait()

	// 3. 此时所有 worker 已退出，取消上下文以清理可能的资源
	n.cancel()
}
