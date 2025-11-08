package yaml

import (
	"encoding/json"
	"strings"
)

func Unmarshal(data []byte, out interface{}) error {
	lines := strings.Split(string(data), "\n")
	result := map[string]map[string]string{}
	var current string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !strings.HasPrefix(line, " ") {
			if strings.HasSuffix(trimmed, ":") {
				current = strings.TrimSuffix(trimmed, ":")
				if result[current] == nil {
					result[current] = map[string]string{}
				}
				continue
			}
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				value = strings.Trim(value, "\"")
				if result[""] == nil {
					result[""] = map[string]string{}
				}
				result[""][key] = value
			}
			continue
		}
		if current == "" {
			continue
		}
		parts := strings.SplitN(strings.TrimSpace(trimmed), ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"")
		if result[current] == nil {
			result[current] = map[string]string{}
		}
		result[current][key] = value
	}
	asJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return json.Unmarshal(asJSON, out)
}
