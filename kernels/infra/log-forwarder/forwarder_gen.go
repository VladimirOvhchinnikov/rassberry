package logforwarder

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"example.com/ffp/platform/ports"
	rt "example.com/ffp/platform/runtime"
	"example.com/ffp/platform/telemetry"
	telemetrypb "example.com/ffp/platform/telemetry/proto"
)

// Forwarder подписывается на локальный EventBus и шлёт логи в Root LogGateway.
type Forwarder struct {
	addr    string
	bus     ports.EventBus
	logger  ports.Logger
	topics  []string
	backoff rt.BackoffPolicy
}

type Option func(*Forwarder)

// WithTopics переопределяет список тем (по умолчанию — все telemetry.logs.* для домена/функций).
func WithTopics(t []string) Option {
	return func(f *Forwarder) {
		if len(t) > 0 {
			f.topics = append([]string(nil), t...)
		}
	}
}

// WithBackoff задаёт стратегию reconnect.
func WithBackoff(b rt.BackoffPolicy) Option { return func(f *Forwarder) { f.backoff = b } }

// New создаёт форвардер. addr — адрес gRPC LogGateway (например, "127.0.0.1:8079").
func New(addr string, bus ports.EventBus, logger ports.Logger, opts ...Option) *Forwarder {
	f := &Forwarder{
		addr:   addr,
		bus:    bus,
		logger: logger,
		topics: []string{
			rt.TopicTelemetryLogsAll,      // общий
			rt.TopicTelemetryLogsDomain,   // доменные
			rt.TopicTelemetryLogsFunction, // функциональные
		},
		backoff: rt.BackoffPolicy{Min: 200 * time.Millisecond, Max: 10 * time.Second, Factor: 2.0, Jitter: 0.2},
	}
	for _, o := range opts {
		o(f)
	}
	return f
}

// Run запускает вечный цикл: connect → stream → reconnect при ошибке.
// At-most-once: при ошибках отправки сообщения теряются (умышленно, без ретраев).
func (f *Forwarder) Run(ctx context.Context) error {
	if f.bus == nil || f.addr == "" {
		return errors.New("forwarder: bus or addr missing")
	}

	// объединяем несколько подписок в один входной канал
	type anyMsg = any
	in := make(chan anyMsg, 256)
	cancels := make([]func(), 0, len(f.topics))
	for _, t := range f.topics {
		ch, cancel, _ := f.bus.Subscribe(ctx, t)
		cancels = append(cancels, cancel)
		go func(c <-chan anyMsg) {
			for m := range c {
				select {
				case in <- m:
				default:
					// drop при переполнении
				}
			}
		}(ch)
	}
	defer func() {
		for _, c := range cancels {
			if c != nil {
				c()
			}
		}
	}()

	attempt := 0
	for {
		attempt++
		// dial
		dctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		conn, err := grpc.DialContext(dctx, f.addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		cancel()
		if err != nil {
			sleep := f.backoff.Duration(attempt)
			f.log(ctx, "WARN", "log-forwarder dial failed", map[string]any{"addr": f.addr, "err": err.Error(), "sleep": sleep.String()})
			select {
			case <-time.After(sleep):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		attempt = 0 // сбрасываем backoff после успешного диала

		client := telemetrypb.NewLogGatewayClient(conn)
		stream, err := client.PushLogs(ctx)
		if err != nil {
			_ = conn.Close()
			sleep := f.backoff.Duration(1)
			f.log(ctx, "WARN", "log-forwarder stream open failed", map[string]any{"err": err.Error(), "sleep": sleep.String()})
			select {
			case <-time.After(sleep):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// основная петля: читаем из локального EventBus и шлём в gRPC
		sendErr := make(chan error, 1)
		go func() {
			defer close(sendErr)
			for {
				select {
				case <-ctx.Done():
					_ = stream.CloseSend()
					return
				case m := <-in:
					if m == nil {
						continue
					}
					rec, ok := m.(telemetry.LogRecordV2)
					if !ok {
						continue // игнорируем другие типы
					}
					// простая защита от циклов: root-логи не пересылаем обратно в root
					if rec.Scope == "root" || rec.KernelID == "rk" {
						continue
					}
					if err := stream.Send(telemetry.ToProto(rec)); err != nil {
						sendErr <- err
						return
					}
				}
			}
		}()

		// ждём ошибки отправки или отмены контекста
		select {
		case <-ctx.Done():
			_ = conn.Close()
			return ctx.Err()
		case err := <-sendErr:
			_ = conn.Close()
			if err == nil {
				return nil
			}
			// reconnect с backoff
			attempt++
			sleep := f.backoff.Duration(attempt)
			f.log(ctx, "WARN", "log-forwarder send failed", map[string]any{"err": err.Error(), "sleep": sleep.String()})
			select {
			case <-time.After(sleep):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func (f *Forwarder) log(ctx context.Context, level, msg string, fields map[string]any) {
	if f.logger != nil {
		f.logger.Log(ctx, level, msg, fields)
	}
}

// Duration — публичная обёртка для BackoffPolicy (удобно в логах/тестах).
func (b BackoffPolicy) Duration(attempt int) time.Duration { return b.duration(attempt) }
