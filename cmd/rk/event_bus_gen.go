package main

import (
	"context"
	"sync"
)

type subscriber struct {
	ch     chan any
	closed bool
}

type InMemoryEventBus struct {
	mu   sync.RWMutex
	subs map[string]map[int]*subscriber
	next int
}

func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{subs: make(map[string]map[int]*subscriber)}
}

func (b *InMemoryEventBus) Publish(ctx context.Context, topic string, msg any) error {
	b.mu.RLock()
	subs := b.subs[topic]
	if len(subs) == 0 {
		b.mu.RUnlock()
		return nil
	}
	// копируем, чтобы не держать блокировку во время отправки
	copies := make([]*subscriber, 0, len(subs))
	for _, sub := range subs {
		copies = append(copies, sub)
	}
	b.mu.RUnlock()
	for _, sub := range copies {
		select {
		case sub.ch <- msg:
		default:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (b *InMemoryEventBus) Subscribe(ctx context.Context, topic string) (<-chan any, func(), error) {
	ch := make(chan any, 16)
	sub := &subscriber{ch: ch}

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.subs[topic] == nil {
		b.subs[topic] = make(map[int]*subscriber)
	}
	id := b.next
	b.next++
	b.subs[topic][id] = sub

	cancel := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if subs, ok := b.subs[topic]; ok {
			if s, ok := subs[id]; ok && !s.closed {
				close(s.ch)
				s.closed = true
				delete(subs, id)
			}
			if len(subs) == 0 {
				delete(b.subs, topic)
			}
		}
	}

	go func() {
		<-ctx.Done()
		cancel()
	}()

	return ch, cancel, nil
}
