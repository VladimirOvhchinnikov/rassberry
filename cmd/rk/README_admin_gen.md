# Root-Kernel (каркас, без перезаписи существующего main.go)

Этот каркас добавляет:
- загрузку YAML-конфига,
- инициализацию DI (tee-логгер, in-proc EventBus, RPC-заглушка),
- admin HTTP (`/admin/health`, `/admin/kernels`).

## Запуск без изменения существующего main.go
Используйте build tag `rk_run`:
```bash
go run -tags rk_run ./cmd/rk -config ./cmd/rk/config.sample.yaml
# или
go build -tags rk_run -o bin/rk ./cmd/rk && ./bin/rk -config ./cmd/rk/config.sample.yaml
```
