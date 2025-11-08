package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type StdLogger struct {
	mu     sync.Mutex
	prefix string
	out    *os.File
}

func NewStdLogger(prefix string) *StdLogger {
	return &StdLogger{prefix: prefix, out: os.Stdout}
}

func (l *StdLogger) Log(_ context.Context, level string, message string, fields map[string]any) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	payload := ""
	if len(fields) > 0 {
		if b, err := json.Marshal(fields); err == nil {
			payload = string(b)
		}
	}
	fmt.Fprintf(l.out, "%s %-5s %s %s\n", time.Now().Format(time.RFC3339), level, l.prefix, message)
	if payload != "" {
		fmt.Fprintln(l.out, "  ", payload)
	}
}
