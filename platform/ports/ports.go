package ports

import "context"

// RPC описывает общий RPC-порт (HTTP/gRPC/JSON-RPC — неважно).
type RPC interface {
	// Register регистрирует обработчик/сервис.
	Register(service any) error
	// Start запускает RPC-слой и блокируется до завершения контекста.
	Start(ctx context.Context) error
}

// EventBus — лёгкий pub/sub порт (at-most-once, in-proc по умолчанию).
type EventBus interface {
	Publish(ctx context.Context, topic string, msg any) error
	// Subscribe возвращает канал сообщений и функцию отмены подписки.
	Subscribe(ctx context.Context, topic string) (<-chan any, func(), error)
}

// Stream — durable-очередь (at-least-once), абстракция над брокерами.
type Stream interface {
	Publish(ctx context.Context, topic string, msg []byte) error
	Consume(ctx context.Context, group, topic string) (<-chan []byte, func(), error)
}

// Logger — минимальный лог-порт для унификации.
type Logger interface {
	Log(ctx context.Context, level string, message string, fields map[string]any)
}
