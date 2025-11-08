# Fractal Platform — монорепозиторий (Go 1.22+)

Этот репозиторий — каркас для фрактальной платформы (Root/Domain/Function Kernels). Шаг №1: подготовлен монорепо на Go 1.22+ с `go.work`, базовыми модулями и минимальными бинарями `rk` и `rkctl`.

## Требования
- Go **1.22+** (`go version` должен показать >=1.22)

## Структура


go.work
Makefile
platform/
contracts/ # Общие контракты (Manifest, Health, Envelope)
ports/ # Порты/интерфейсы (RPC, EventBus, Stream, Logger)
runtime/ # Вспомогательное рантайм-API (заглушки)
telemetry/ # Типы для логов (Level, LogRecord)
cmd/
rk/ # Скелет Root-Kernel (main)
rkctl/ # Скелет CLI (main)


## Быстрый старт
```bash
make tidy
make build
./bin/rk
./bin/rkctl -version
```

Либо:

```
go run ./cmd/rk
go run ./cmd/rkctl
```

Смена модульного префикса

Сейчас используется example.com/ffp. Позже замените на свой:

Во всех go.mod и исходниках замените example.com/ffp на gitlab.com/<group>/<repo> (или другой).

Запустите make tidy.

Что дальше

На следующем шаге будут добавлены супервизор, discovery, admin-API и поток логов.
