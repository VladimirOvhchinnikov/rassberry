package telemetry

// LogFilter — общий фильтр по полям лога.
// Поле не участвует в фильтрации, если пустое/нулевое.
type LogFilter struct {
	LevelMin  Level  // пропускать записи >= LevelMin (если LevelMin=0 и это Debug — считается заданным)
	KernelID  string // точное совпадение, если задано
	Scope     string // точное совпадение, если задано
	Component string // префиксное совпадение, если задано (например "gateway" матчит "gateway/http")
}

// Match сообщает, проходит ли запись фильтр.
func (f LogFilter) MatchV1(r struct {
	Level     Level
	KernelID  string
	Scope     string
	Component string
}) bool {
	if f.KernelID != "" && r.KernelID != f.KernelID {
		return false
	}
	if f.Scope != "" && r.Scope != f.Scope {
		return false
	}
	if f.Component != "" {
		if r.Component == "" {
			return false
		}
		if len(r.Component) < len(f.Component) || r.Component[:len(f.Component)] != f.Component {
			return false
		}
	}
	// Уровень: если LevelMin задан, сравниваем
	// Замечание: нулевой Level может означать Debug (0); поэтому всегда сравниваем.
	return r.Level >= f.LevelMin
}

// MatchV2 — фильтрация LogRecordV2.
func (f LogFilter) MatchV2(r LogRecordV2) bool {
	return f.MatchV1(struct {
		Level     Level
		KernelID  string
		Scope     string
		Component string
	}{
		Level: r.Level, KernelID: r.KernelID, Scope: r.Scope, Component: r.Component,
	})
}
