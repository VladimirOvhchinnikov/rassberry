# DK: examples/site

- Поднимает HTTP `GET /hello` (адрес `http_addr`, по умолчанию `:8081`).
- Включает FK `log-forwarder`: пересылает логи в Root LogGateway (`log_gateway`, по умолчанию `127.0.0.1:8079`).
- Фоновый воркер пишет heartbeat-логи каждые 3с.

## Конфиг домена
```yaml
domains:
  - id: "site"
    mode: "inproc"
    kind: "site"
    config:
      http_addr: ":8081"
      log_gateway: "127.0.0.1:8079"
```

Примечание: подключение kind: "site" к лаунчеру — отдельный шаг (в DomainKernelLauncher добавить case "site": kernel = site.NewDomain(spec.ID)).
