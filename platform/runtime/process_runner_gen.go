package runtime

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

// LogLine — единичная строка лога процесса.
type LogLine struct {
	Time   time.Time
	Origin string // "stdout" | "stderr"
	Line   string
}

// ProcEventType — тип события процесса.
type ProcEventType string

const (
	ProcStart   ProcEventType = "start"
	ProcExit    ProcEventType = "exit"
	ProcReady   ProcEventType = "ready"
	ProcProbeOK ProcEventType = "probe_ok"
	ProcProbeKO ProcEventType = "probe_ko"
)

// ProcEvent — событие жизненного цикла процесса.
type ProcEvent struct {
	Time     time.Time
	Type     ProcEventType
	ExitCode int
	Err      error
	Note     string
}

// ProcessRunner — каркас для запуска внешнего DK-процесса.
type ProcessRunner struct {
	cmdPath     string
	args        []string
	env         []string
	wd          string
	healthURL   string
	readyPrefix string

	logCh   chan LogLine
	onEvent func(ProcEvent)

	mu  sync.Mutex
	cmd *exec.Cmd

	ready bool
}

// BackoffPolicy — параметры экспоненциального backoff с джиттером.
type BackoffPolicy struct {
	Min    time.Duration
	Max    time.Duration
	Factor float64
	Jitter float64
}

func (b BackoffPolicy) withDefaults() BackoffPolicy {
	out := b
	if out.Min <= 0 {
		out.Min = 100 * time.Millisecond
	}
	if out.Max <= 0 {
		out.Max = 30 * time.Second
	}
	if out.Factor <= 0 {
		out.Factor = 2.0
	}
	if out.Jitter < 0 {
		out.Jitter = 0
	}
	if out.Jitter > 1 {
		out.Jitter = 1
	}
	return out
}

func (b BackoffPolicy) duration(attempt int) time.Duration {
	b = b.withDefaults()
	// exp := Min * Factor^(attempt-1)
	exp := float64(b.Min)
	for i := 1; i < attempt; i++ {
		exp *= b.Factor
	}
	d := time.Duration(exp)
	if d > b.Max {
		d = b.Max
	}
	// лёгкий джиттер
	if b.Jitter > 0 {
		// используем текущее время как источник случайности
		n := time.Now().UnixNano()
		j := float64(int64(n%2000)-1000) / 1000.0 // [-1; +1] приблизительно
		delta := time.Duration(float64(d) * j * b.Jitter)
		d += delta
		if d < b.Min {
			d = b.Min
		}
		if d > b.Max {
			d = b.Max
		}
	}
	return d
}

// NewProcessRunner создаёт новый раннер.
func NewProcessRunner(cmdPath string, args []string, opts ...PROption) *ProcessRunner {
	pr := &ProcessRunner{
		cmdPath:     cmdPath,
		args:        append([]string(nil), args...),
		readyPrefix: "READY",
		logCh:       make(chan LogLine, 256),
	}
	for _, o := range opts {
		o(pr)
	}
	return pr
}

type PROption func(*ProcessRunner)

func WithEnvMap(m map[string]string) PROption {
	return func(p *ProcessRunner) {
		if len(m) == 0 {
			return
		}
		for k, v := range m {
			p.env = append(p.env, k+"="+v)
		}
	}
}
func WithWorkingDir(dir string) PROption { return func(p *ProcessRunner) { p.wd = dir } }
func WithHealthHTTP(url string) PROption { return func(p *ProcessRunner) { p.healthURL = url } }
func WithReadyPrefix(prefix string) PROption {
	return func(p *ProcessRunner) {
		if prefix != "" {
			p.readyPrefix = prefix
		}
	}
}
func WithOnEvent(h func(ProcEvent)) PROption { return func(p *ProcessRunner) { p.onEvent = h } }

// Logs возвращает канал строк логов (stdout/stderr).
func (p *ProcessRunner) Logs() <-chan LogLine { return p.logCh }

func (p *ProcessRunner) emit(e ProcEvent) {
	if p.onEvent != nil {
		defer func() { _ = recover() }()
		p.onEvent(e)
	}
}

// Start запускает процесс и читает stdout/stderr в фоне.
func (p *ProcessRunner) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cmd != nil {
		return errors.New("process already started")
	}
	if p.cmdPath == "" {
		return errors.New("empty command path")
	}
	cmd := exec.CommandContext(ctx, p.cmdPath, p.args...)
	if p.wd != "" {
		cmd.Dir = p.wd
	}
	if len(p.env) > 0 {
		cmd.Env = append(os.Environ(), p.env...)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	p.cmd = cmd
	p.emit(ProcEvent{Time: time.Now(), Type: ProcStart, Note: filepath.Base(p.cmdPath)})

	// читалки stdout/stderr
	go p.readPipe("stdout", stdout)
	go p.readPipe("stderr", stderr)

	// наблюдаем за завершением
	go func() {
		err := cmd.Wait()
		code := 0
		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok && ee.ProcessState != nil {
				code = ee.ProcessState.ExitCode()
			} else {
				code = -1
			}
		}
		p.emit(ProcEvent{Time: time.Now(), Type: ProcExit, ExitCode: code, Err: err})
		close(p.logCh)
	}()

	return nil
}

func (p *ProcessRunner) readPipe(origin string, r io.ReadCloser) {
	defer r.Close()
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		p.logCh <- LogLine{Time: time.Now(), Origin: origin, Line: line}
		if origin == "stdout" && p.readyPrefix != "" && strings.HasPrefix(line, p.readyPrefix) {
			p.ready = true
			p.emit(ProcEvent{Time: time.Now(), Type: ProcReady, Note: line})
		}
	}
}

func (p *ProcessRunner) Ready() bool { return p.ready }

// WaitReady ожидает готовности по stdout (READY) или по HTTP-пробе, что наступит раньше.
func (p *ProcessRunner) WaitReady(ctx context.Context, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.New("wait ready: timeout or cancelled")
		case <-ticker.C:
			if p.ready {
				return nil
			}
			if p.healthURL != "" {
				cl := &http.Client{Timeout: 1 * time.Second}
				req, _ := http.NewRequestWithContext(ctx, http.MethodGet, p.healthURL, nil)
				resp, err := cl.Do(req)
				if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
					p.emit(ProcEvent{Time: time.Now(), Type: ProcProbeOK, Note: p.healthURL})
					return nil
				}
				p.emit(ProcEvent{Time: time.Now(), Type: ProcProbeKO, Note: p.healthURL, Err: err})
			}
		}
	}
}

// Stop посылает мягкий сигнал завершения процессу.
func (p *ProcessRunner) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	if runtime.GOOS == "windows" {
		return p.cmd.Process.Kill()
	}
	return p.cmd.Process.Signal(syscall.SIGTERM)
}
