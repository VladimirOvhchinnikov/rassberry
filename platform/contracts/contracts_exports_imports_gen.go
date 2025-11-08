package contracts

// ВНИМАНИЕ: Автогенерируемый файл-дополнение.
// Он добавляет недостающие модели Exports/Imports, не изменяя существующие типы и комментарии.

// NetworkEndpoint описывает экспортируемую сетевую точку ядра (HTTP/gRPC/TCP).
type NetworkEndpoint struct {
	// Логическое имя сервиса/ручки (например, "gateway", "users").
	Name string `json:"name" yaml:"name"`
	// Протокол: "http" | "grpc" | "tcp".
	Protocol string `json:"protocol" yaml:"protocol"`
	// Адрес/путь: для http — base-path или host:port, для grpc/tcp — host:port.
	Address string `json:"address" yaml:"address"`
	// Версия API/контракта, например "v1".
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	// Список путей/методов (для HTTP) либо методов сервиса (для gRPC) — опционально.
	Endpoints []string `json:"endpoints,omitempty" yaml:"endpoints,omitempty"`
}

// EventSpec описывает экспортируемую тему лёгкого шина-событий (in-proc bus).
type EventSpec struct {
	Topic       string `json:"topic" yaml:"topic"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// StreamSpec описывает экспортируемую тему стрима/брокера (durable).
type StreamSpec struct {
	Topic    string `json:"topic" yaml:"topic"`
	Group    string `json:"group,omitempty" yaml:"group,omitempty"`
	DLQ      string `json:"dlq,omitempty" yaml:"dlq,omitempty"`
	Delivery string `json:"delivery,omitempty" yaml:"delivery,omitempty"` // например: "at-least-once"
}

// CLICommand описывает экспортируемую CLI/админ-команду.
type CLICommand struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// LocalService описывает локально предоставляемый интерфейс внутри процесса/домена.
type LocalService struct {
	Name      string `json:"name" yaml:"name"`
	Interface string `json:"interface" yaml:"interface"`
	Version   string `json:"version,omitempty" yaml:"version,omitempty"`
}

// Exports агрегирует всё, что ядро предоставляет наружу.
type Exports struct {
	Network []NetworkEndpoint `json:"network,omitempty" yaml:"network,omitempty"`
	Events  []EventSpec       `json:"events,omitempty" yaml:"events,omitempty"`
	Streams []StreamSpec      `json:"streams,omitempty" yaml:"streams,omitempty"`
	CLI     []CLICommand      `json:"cli,omitempty" yaml:"cli,omitempty"`
	Local   []LocalService    `json:"local,omitempty" yaml:"local,omitempty"`
}

// RPCRef — ссылка на внешние RPC-сервисы, от которых зависит ядро.
type RPCRef struct {
	Name     string `json:"name" yaml:"name"` // например: svc://site.gateway@v1
	Version  string `json:"version,omitempty" yaml:"version,omitempty"`
	Optional bool   `json:"optional,omitempty" yaml:"optional,omitempty"`
}

// TopicRef — подписка на темы событий (лёгкий bus).
type TopicRef struct {
	Topic    string `json:"topic" yaml:"topic"`
	Optional bool   `json:"optional,omitempty" yaml:"optional,omitempty"`
}

// StreamRef — подписка на темы стрима/брокера.
type StreamRef struct {
	Topic    string `json:"topic" yaml:"topic"`
	Group    string `json:"group,omitempty" yaml:"group,omitempty"`
	Optional bool   `json:"optional,omitempty" yaml:"optional,omitempty"`
}

// StorageRef — зависимость от хранилищ/клиентов (kv/sql/blob и т.п.).
type StorageRef struct {
	Kind     string `json:"kind" yaml:"kind"` // пример: "sql" | "kv" | "blob"
	Name     string `json:"name" yaml:"name"` // логическое имя
	Optional bool   `json:"optional,omitempty" yaml:"optional,omitempty"`
}

// Imports агрегирует всё, что ядро ожидает от окружения.
type Imports struct {
	RPC      []RPCRef     `json:"rpc,omitempty" yaml:"rpc,omitempty"`
	Events   []TopicRef   `json:"events,omitempty" yaml:"events,omitempty"`
	Streams  []StreamRef  `json:"streams,omitempty" yaml:"streams,omitempty"`
	Storages []StorageRef `json:"storages,omitempty" yaml:"storages,omitempty"`
	Env      []string     `json:"env,omitempty" yaml:"env,omitempty"` // имена переменных окружения
}
