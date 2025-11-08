# DomainKernelLauncher

Лаунчер поддерживает режимы `inproc`, `process` (заглушка) и `remote` (возвращает ошибку).

## DomainSpec
- `id` — идентификатор доменного ядра.
- `mode` — `inproc` | `process` | `remote`.
- `kind` — тип домена. Встроенный пример: `example`.
- `config` — произвольный конфиг, пробрасывается в `KernelHost`.

## Расширение
Чтобы подключить реальный домен, замените ветку `switch spec.Kind` в `launchInproc` и верните фабрику, создающую ваш `rt.KernelModule`.

## Статусы для process/remote
- `process` — пока регистрируется как placeholder (`HealthDegraded`).
- `remote` — не реализован, возвращает ошибку.

## Запуск
```bash
go work sync
go run -tags rk_run ./cmd/rk -config ./cmd/rk/config.sample.yaml
```
