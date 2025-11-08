package runtime

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// RestartPolicy описывает стратегию перезапуска воркера.
type RestartPolicy int

const (
	Permanent RestartPolicy = iota // перезапускать всегда (даже при успешном завершении)
	Transient                      // перезапускать только при ошибке/панике
	Temporary                      // не перезапускать никогда
)

func (p RestartPolicy) String() string {
	switch p {
	case Permanent:
		return "permanent"
	case Transient:
		return "transient"
	case Temporary:
		return "temporary"
	default:
		return "unknown"
	}
}

// WorkerFunc — функция воркера. Должна завершиться сама или по отмене ctx.
type WorkerFunc func(ctx context.Context) error

// BackoffPolicy — параметры экспоненциального backoff с джиттером.
type BackoffPolicy struct {
	Min    time.Duration // минимальная задержка (по умолчанию 100ms)
	Max    time.Duration // максимальная задержка (по умолчанию 30s)
	Factor float64       // множитель экспоненты (по умолчанию 2.0)
	Jitter float64       // 0..1, доля случайного джиттера (по умолчанию 0.2)
}

func (b BackoffPolicy) withDefaults() BackoffPolicy {
	out := b
	if out.Min <= 0 {
		out.Min = 100 * time.Millisecond
	}
	if out.Max <= 0 {
		out.Max = 30 * time.Second
	}
	if out.Factor <= 0 {
		out.Factor = 2.0
	}
	if out.Jitter < 0 {
		out.Jitter = 0
	}
	if out.Jitter > 1 {
		out.Jitter = 1
	}
	return out
}

func (b BackoffPolicy) duration(attempt int, rnd *rand.Rand) time.Duration {
	b = b.withDefaults()
	// exp := Min * Factor^(attempt-1)
	exp := float64(b.Min) * pow(b.Factor, attempt-1)
	d := time.Duration(exp)
	if d > b.Max {
		d = b.Max
	}
	// джиттер +/- Jitter*50%
	if b.Jitter > 0 {
		j := (rnd.Float64()*2 - 1) * b.Jitter // [-Jitter, +Jitter]
		delta := time.Duration(float64(d) * j)
		d += delta
		if d < b.Min {
			d = b.Min
		}
		if d > b.Max {
			d = b.Max
		}
	}
	return d
}

func pow(a float64, n int) float64 {
	out := 1.0
	for i := 0; i < n; i++ {
		out *= a
	}
	return out
}

// WorkerSpec описывает запуск единичного воркера.
type WorkerSpec struct {
	Name    string
	Policy  RestartPolicy
	Backoff BackoffPolicy
	Fn      WorkerFunc
}

// EventType — тип события супервизора.
type EventType string

const (
	EventStart   EventType = "start"
	EventExit    EventType = "exit"
	EventRestart EventType = "restart"
	EventPanic   EventType = "panic"
	EventStop    EventType = "stop"
)

// Event — событие жизненного цикла воркера.
type Event struct {
	Time      time.Time
	Worker    string
	Type      EventType
	Attempt   int
	Err       error
	NextAfter time.Duration
}

// Supervisor — надзор за воркерами с политиками рестартов.
type Supervisor struct {
	mu      sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	workers map[string]*worker
	onEvent func(Event)

	rnd *rand.Rand
}

type worker struct {
	spec     WorkerSpec
	ctx      context.Context
	cancel   context.CancelFunc
	attempts int
	stopped  bool
}

type SupervisorOption func(*Supervisor)

// WithOnEvent задаёт обработчик событий (безопасно к панике пользователя).
func WithOnEvent(h func(Event)) SupervisorOption {
	return func(s *Supervisor) { s.onEvent = h }
}

// NewSupervisor создаёт новый супервизор с собственным контекстом.
func NewSupervisor(opts ...SupervisorOption) *Supervisor {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Supervisor{
		ctx:     ctx,
		cancel:  cancel,
		workers: make(map[string]*worker),
		rnd:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

func (s *Supervisor) emit(e Event) {
	if s.onEvent == nil {
		return
	}
	defer func() { _ = recover() }()
	s.onEvent(e)
}

// Start запускает воркер по спецификации.
func (s *Supervisor) Start(spec WorkerSpec) error {
	if spec.Fn == nil || spec.Name == "" {
		return fmt.Errorf("invalid worker spec")
	}
	s.mu.Lock()
	if _, exists := s.workers[spec.Name]; exists {
		s.mu.Unlock()
		return fmt.Errorf("worker %q already exists", spec.Name)
	}
	wctx, cancel := context.WithCancel(s.ctx)
	w := &worker{spec: spec, ctx: wctx, cancel: cancel}
	s.workers[spec.Name] = w
	s.wg.Add(1)
	s.mu.Unlock()

	go s.runWorker(w)
	return nil
}

// Stop останавливает воркер по имени (мягкая остановка).
func (s *Supervisor) Stop(name string) {
	s.mu.Lock()
	w, ok := s.workers[name]
	if ok && !w.stopped {
		w.stopped = true
		w.cancel()
	}
	s.mu.Unlock()
}

// StopAll останавливает все воркеры и сам супервизор.
func (s *Supervisor) StopAll() {
	s.cancel()
	s.mu.Lock()
	for _, w := range s.workers {
		if !w.stopped {
			w.stopped = true
			w.cancel()
		}
	}
	s.mu.Unlock()
}

// Wait ожидает завершения всех воркеров.
func (s *Supervisor) Wait() { s.wg.Wait() }

func (s *Supervisor) runWorker(w *worker) {
	defer s.wg.Done()
	b := w.spec.Backoff.withDefaults()

	for {
		select {
		case <-s.ctx.Done():
			s.emit(Event{Time: time.Now(), Worker: w.spec.Name, Type: EventStop})
			return
		default:
		}

		s.emit(Event{Time: time.Now(), Worker: w.spec.Name, Type: EventStart, Attempt: w.attempts + 1})

		err := func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic: %v", r)
					s.emit(Event{Time: time.Now(), Worker: w.spec.Name, Type: EventPanic, Attempt: w.attempts + 1, Err: err})
				}
			}()
			return w.spec.Fn(w.ctx)
		}()

		w.attempts++
		s.emit(Event{Time: time.Now(), Worker: w.spec.Name, Type: EventExit, Attempt: w.attempts, Err: err})

		// Решение о рестарте по политике
		restart := false
		switch w.spec.Policy {
		case Permanent:
			restart = true
		case Transient:
			restart = err != nil
		case Temporary:
			restart = false
		}

		if !restart {
			// завершить и удалить
			s.mu.Lock()
			delete(s.workers, w.spec.Name)
			s.mu.Unlock()
			return
		}

		// backoff + jitter
		sleep := b.duration(w.attempts, s.rnd)
		s.emit(Event{Time: time.Now(), Worker: w.spec.Name, Type: EventRestart, Attempt: w.attempts, Err: err, NextAfter: sleep})

		select {
		case <-time.After(sleep):
		case <-s.ctx.Done():
			return
		}
	}
}
