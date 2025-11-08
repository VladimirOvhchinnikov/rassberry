package telemetry

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// PublisherFunc — внешний паблишер сообщений (например, обёртка над EventBus.Publish).
type PublisherFunc func(ctx context.Context, topic string, msg any) error

type SafeLogHubOption func(*SafeLogHub)

func WithTopics(topics ...string) SafeLogHubOption {
	return func(h *SafeLogHub) { h.topics = append([]string(nil), topics...) }
}
func WithScopeTopic(enable bool) SafeLogHubOption {
	return func(h *SafeLogHub) { h.scopeTopic = enable }
}
func WithBuffer(size int) SafeLogHubOption {
	return func(h *SafeLogHub) {
		if size > 0 {
			h.buf = size
		}
	}
}
func WithRateLimit(maxPerInterval int, interval time.Duration) SafeLogHubOption {
	return func(h *SafeLogHub) {
		if maxPerInterval <= 0 {
			maxPerInterval = 1000
		}
		if interval <= 0 {
			interval = time.Second
		}
		h.rateMax = int64(maxPerInterval)
		h.rateWindow = interval
	}
}

// SafeLogHub — защищённый хаб логов (очередь + rate-limit + счётчики).
type SafeLogHub struct {
	pub PublisherFunc

	topics     []string
	scopeTopic bool
	buf        int

	// rate-limit
	rateMax    int64
	rateWindow time.Duration
	rateNow    int64
	rateMu     sync.Mutex
	rateTick   *time.Ticker
	lastReset  time.Time

	// очередь
	in chan LogRecordV2

	// счётчики
	total        uint64
	forwarded    uint64
	droppedRate  uint64
	droppedQueue uint64

	stop chan struct{}
	wg   sync.WaitGroup
}

func NewSafeLogHub(pub PublisherFunc, opts ...SafeLogHubOption) *SafeLogHub {
	h := &SafeLogHub{
		pub:        pub,
		topics:     []string{"telemetry.logs"},
		buf:        1024,
		rateMax:    1000,
		rateWindow: time.Second,
		stop:       make(chan struct{}),
	}
	for _, o := range opts {
		o(h)
	}
	h.in = make(chan LogRecordV2, h.buf)
	h.rateTick = time.NewTicker(h.rateWindow)
	h.lastReset = time.Now()

	h.wg.Add(1)
	go h.loop()
	return h
}

func (h *SafeLogHub) loop() {
	defer h.wg.Done()
	for {
		select {
		case <-h.stop:
			return
		case <-h.rateTick.C:
			atomic.StoreInt64(&h.rateNow, 0)
			h.lastReset = time.Now()
		case rec := <-h.in:
			h.forward(rec)
		}
	}
}

func (h *SafeLogHub) forward(rec LogRecordV2) {
	for _, t := range h.topics {
		_ = h.pub(context.Background(), t, rec)
	}
	if h.scopeTopic && rec.Scope != "" {
		_ = h.pub(context.Background(), "telemetry.logs."+rec.Scope, rec)
	}
	atomic.AddUint64(&h.forwarded, 1)
}

// Publish — попытка принять запись. Возвращает true, если принял; false, если дропнул.
func (h *SafeLogHub) Publish(_ context.Context, rec LogRecordV2) bool {
	atomic.AddUint64(&h.total, 1)

	// rate-limit
	if h.rateMax > 0 {
		if atomic.AddInt64(&h.rateNow, 1) > h.rateMax {
			atomic.AddUint64(&h.droppedRate, 1)
			return false
		}
	}

	// неблокирующая запись в очередь
	select {
	case h.in <- rec:
		return true
	default:
		atomic.AddUint64(&h.droppedQueue, 1)
		return false
	}
}

type LogHubStats struct {
	Total        uint64        `json:"total"`
	Forwarded    uint64        `json:"forwarded"`
	DroppedRate  uint64        `json:"dropped_rate"`
	DroppedQueue uint64        `json:"dropped_queue"`
	RateWindow   time.Duration `json:"rate_window"`
	RateMax      int64         `json:"rate_max"`
	Since        time.Time     `json:"since"`
}

func (h *SafeLogHub) Stats() LogHubStats {
	return LogHubStats{
		Total:        atomic.LoadUint64(&h.total),
		Forwarded:    atomic.LoadUint64(&h.forwarded),
		DroppedRate:  atomic.LoadUint64(&h.droppedRate),
		DroppedQueue: atomic.LoadUint64(&h.droppedQueue),
		RateWindow:   h.rateWindow,
		RateMax:      h.rateMax,
		Since:        h.lastReset,
	}
}

func (h *SafeLogHub) Close() {
	close(h.stop)
	h.wg.Wait()
	h.rateTick.Stop()
}
