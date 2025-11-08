# FK: log-forwarder

- Подписывается на локальный `EventBus` (темы `telemetry.logs`, `telemetry.logs.domain`, `telemetry.logs.function`).
- Отправляет записи в Root через gRPC `LogGateway` (bi-di stream).
- Reconnect с экспоненциальным backoff + jitter (без ретраев отдельных сообщений — at-most-once).
- Защита от петель: логи `scope=root`/`kernel_id=rk` не пересылаются.
