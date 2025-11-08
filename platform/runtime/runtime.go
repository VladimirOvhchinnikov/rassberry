package runtime

import "runtime"

// GoRuntimeVersion возвращает версию рантайма Go (для отладки/логов).
func GoRuntimeVersion() string {
	return runtime.Version()
}
