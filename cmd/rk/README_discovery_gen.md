# DiscoveryRegistry (in-memory)

`DiscoveryRegistry` хранит записи о зарегистрированных доменных и функциональных ядрах.

## Структуры
- `KernelRecord` — информация о ядре (манифест, экспорт, health, время регистрации).
- `DiscoveryRegistry` — потокобезопасное хранилище `id -> KernelRecord`.

## Приоритет статусов health
Агрегатор `AggregateHealth` возвращает:
1. `failed`, если хоть одно ядро в состоянии failure.
2. `degraded`, если нет failure, но есть degraded.
3. `draining`, если нет предыдущих, но есть draining.
4. `ready` (по умолчанию) или `stopped`, если записей нет — возвращается `ready` через `readyHealth()`.
