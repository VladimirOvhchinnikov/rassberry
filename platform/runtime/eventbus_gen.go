package runtime

import (
	"context"
	"sync"

	"example.com/ffp/platform/ports"
)

type inprocEventBus struct {
	mu     sync.RWMutex
	subs   map[string]map[chan any]struct{}
	buffer int
}

// NewInprocEventBus создаёт потокобезопасную in-proc шину событий с заданным буфером.
func NewInprocEventBus(buffer int) ports.EventBus {
	if buffer <= 0 {
		buffer = 1
	}
	return &inprocEventBus{
		subs:   make(map[string]map[chan any]struct{}),
		buffer: buffer,
	}
}

func (b *inprocEventBus) Publish(ctx context.Context, topic string, msg any) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if topic == "" {
		return nil
	}
	chans := b.subs[topic]
	for ch := range chans {
		select {
		case ch <- msg:
		default:
		}
	}
	return nil
}

func (b *inprocEventBus) Subscribe(ctx context.Context, topic string) (<-chan any, func(), error) {
	if topic == "" {
		topic = "default"
	}
	ch := make(chan any, b.buffer)
	b.mu.Lock()
	if _, ok := b.subs[topic]; !ok {
		b.subs[topic] = make(map[chan any]struct{})
	}
	b.subs[topic][ch] = struct{}{}
	b.mu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			b.mu.Lock()
			if m, ok := b.subs[topic]; ok {
				if _, exists := m[ch]; exists {
					delete(m, ch)
					if len(m) == 0 {
						delete(b.subs, topic)
					}
				}
			}
			b.mu.Unlock()
			close(ch)
		})
	}

	go func() {
		select {
		case <-ctx.Done():
			cancel()
		}
	}()

	return ch, cancel, nil
}
