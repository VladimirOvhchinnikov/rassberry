package telemetry

import "time"

// ВНИМАНИЕ: если в пакете уже есть LogRecord — НЕ переопределяйте его.
// В этом случае добавьте расширенную модель LogRecordV2 (ниже).
// Если LogRecord отсутствует — можно объявить его здесь (с полем Trace).

// Проверочный комментарий для человека: генератор должен проверить наличие типа
// LogRecord в существующих файлах. Если тип уже есть — объявлять ТОЛЬКО LogRecordV2.

// --- ВАРИАНТ A: базовая модель (если LogRecord отсутствует) ---
// РАЗРЕШАЕТСЯ СОЗДАТЬ нижеуказанный тип LogRecord (включая поле Trace).
// Закомментировано, чтобы не спровоцировать дублирование при чтении.
// Раскомментируй этот блок ТОЛЬКО если в пакете нет LogRecord.
/*
type Level int

const (
    Debug Level = iota
    Info
    Warn
    Error
)

type LogRecord struct {
    Time      time.Time      `json:"time"`
    Level     Level          `json:"level"`
    KernelID  string         `json:"kernel_id"`
    Scope     string         `json:"scope"`
    Component string         `json:"component,omitempty"`
    Trace     string         `json:"trace,omitempty"` // W3C traceparent или иной идентификатор
    Message   string         `json:"message"`
    Fields    map[string]any `json:"fields,omitempty"`
}
*/

// --- ВАРИАНТ B: расширенная модель (если LogRecord уже существует) ---
type LogRecordV2 struct {
	Time      time.Time      `json:"time"`
	Level     Level          `json:"level"`
	KernelID  string         `json:"kernel_id"`
	Scope     string         `json:"scope"`
	Component string         `json:"component,omitempty"`
	Trace     string         `json:"trace,omitempty"` // W3C traceparent или иной идентификатор
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
}

// FromV1 формирует V2-запись из совместимой v1-записи (без trace).
// Этот помощник не трогает исходную структуру, чтобы не ломать совместимость.
func FromV1(v1 struct {
	Time      time.Time
	Level     Level
	KernelID  string
	Scope     string
	Component string
	Message   string
	Fields    map[string]any
}) LogRecordV2 {
	return LogRecordV2{
		Time:      v1.Time,
		Level:     v1.Level,
		KernelID:  v1.KernelID,
		Scope:     v1.Scope,
		Component: v1.Component,
		Message:   v1.Message,
		Fields:    v1.Fields,
	}
}
