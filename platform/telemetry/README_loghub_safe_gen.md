# SafeLogHub — защита от лог-штормов

- Ограничение скорости (`WithRateLimit(maxPerInterval, interval)`).
- Ограниченная очередь (`WithBuffer(size)`) — при переполнении записи дропаются.
- Публикация в базовые темы (`WithTopics(...)`) и опционально в `telemetry.logs.<scope>` (`WithScopeTopic(true)`).
- Метрики: `Stats()` — всего/отправлено/дроп по rate/дроп по очереди.

Использование (пример в Root-Kernel):
```go
pub := func(ctx context.Context, topic string, msg any) error { return bus.Publish(ctx, topic, msg) }
hub := telemetry.NewSafeLogHub(pub,
    telemetry.WithTopics("telemetry.logs", "telemetry.logs.root"),
    telemetry.WithScopeTopic(true),
    telemetry.WithRateLimit(2000, time.Second),
    telemetry.WithBuffer(2048),
)
// hub.Publish(ctx, rec) — nonblocking; false = запись отброшена.
```

> Интеграцию этого хаба в текущий gRPC-сервер/LogGateway можно сделать отдельным шагом (через `StartLogGatewayServerSafe(...)` или замену использования на `hub.Publish`). Никаких существующих файлов не перезаписывай.

---

## Напоминание по запуску

После каждого промта:
```bash
go work sync
go build ./...
```

Чтобы активировать деградацию и фабрику site, запусти RK с примерным конфигом:
```bash
go run -tags rk_run ./cmd/rk -config ./cmd/rk/config.sample.yaml
```

Проверка:
```bash
curl http://localhost:8090/admin/kernels     # увидишь запись для "site" с exports
curl http://localhost:8090/admin/health
```

Если переведёшь site в degraded (временно через POST /admin/kernels/site/drain), в ответе /admin/kernels у него пропадут exports до восстановления статуса.
