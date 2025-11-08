//go:build e2e

package e2e

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type logRec struct {
	KernelID string `json:"kernel_id"`
	Scope    string `json:"scope"`
	Level    int    `json:"level"`
	Message  string `json:"message"`
}

func TestLogsStream_FilterChange(t *testing.T) {
	// готовим конфиг во временном файле
	cfg := `
root: { node_id: "rk-1", zone: "dc-1" }
admin: { addr: ":8090", grpc_addr: ":8079" }
discovery: { enabled: true, advertise_internal: true }
telemetry: { level: "INFO", buffer: 256, filters: { level: "INFO" } }
domains:
  - id: "site"
    mode: "inproc"
    kind: "site"
    feature_flags: { http: true, workers: true, log_forwarder: true }
    config: { http_addr: ":8081", log_gateway: "127.0.0.1:8079" }
`
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "rk.yaml")
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// стартуем rk
	cmd := exec.CommandContext(ctx, "go", "run", "-tags", "rk_run", "./cmd/rk", "-config", cfgPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	// ждём готовности admin
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://localhost:8090/admin/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// helper для чтения первой записи из SSE
	readOnce := func(url string, wait time.Duration) (bool, error) {
		client := &http.Client{Timeout: 0}
		resp, err := client.Get(url)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()
		rd := bufio.NewReader(resp.Body)
		timer := time.NewTimer(wait)
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				return false, nil
			default:
				line, err := rd.ReadString('\n')
				if err != nil {
					return false, err
				}
				if !strings.HasPrefix(line, "data: ") {
					continue
				}
				raw := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
				var lr logRec
				if err := json.Unmarshal([]byte(raw), &lr); err == nil {
					return true, nil
				}
			}
		}
	}

	// 1) level=ERROR — ждём 5с: записей не должно быть
	ok, err := readOnce("http://localhost:8090/admin/logs/stream?kernel=site&level=ERROR", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("unexpected logs at level=ERROR")
	}

	// 2) level=DEBUG — ждём до 12с: запись должна прийти
	ok, err = readOnce("http://localhost:8090/admin/logs/stream?kernel=site&level=DEBUG", 12*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("no logs received at level=DEBUG")
	}
}
