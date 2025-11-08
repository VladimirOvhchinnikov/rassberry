//go:build rkctl_run

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"example.com/ffp/platform/telemetry"
)

func usage() {
	fmt.Fprintf(os.Stderr, `rkctl — CLI

Команды:
  rkctl logs [--http URL] [--level L] [--kernel ID] [--scope S] [--component C] [--pretty] [--compact]
  rkctl kernels list   [--http URL]
  rkctl kernels health [--http URL]
  rkctl kernels restart --id ID [--http URL]
  rkctl kernels drain   --id ID [--http URL]

По умолчанию --http=http://localhost:8090
`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}
	switch os.Args[1] {
	case "logs":
		cmdLogs(os.Args[2:])
	case "kernels":
		if len(os.Args) < 3 {
			usage()
			return
		}
		switch os.Args[2] {
		case "list":
			cmdKernelsList(os.Args[3:])
		case "health":
			cmdKernelsHealth(os.Args[3:])
		case "restart":
			cmdKernelsAction(os.Args[3:], "restart")
		case "drain":
			cmdKernelsAction(os.Args[3:], "drain")
		default:
			usage()
		}
	default:
		usage()
	}
}

func defaultHTTP() string { return "http://localhost:8090" }

func cmdLogs(args []string) {
	fs := flag.NewFlagSet("logs", flag.ExitOnError)
	httpURL := fs.String("http", defaultHTTP(), "Base URL admin HTTP")
	level := fs.String("level", "INFO", "Level threshold: DEBUG|INFO|WARN|ERROR")
	kernel := fs.String("kernel", "", "Kernel ID filter")
	scope := fs.String("scope", "", "Scope filter: root|domain|function")
	component := fs.String("component", "", "Component prefix")
	pretty := fs.Bool("pretty", false, "Pretty JSON in payload")
	compact := fs.Bool("compact", false, "Compact one-line format")
	_ = fs.Parse(args)

	q := fmt.Sprintf("%s/admin/logs/stream?level=%s&kernel=%s&scope=%s&component=%s",
		strings.TrimRight(*httpURL, "/"),
		urlQueryEsc(*level), urlQueryEsc(*kernel), urlQueryEsc(*scope), urlQueryEsc(*component),
	)
	resp, err := http.Get(q)
	if err != nil {
		fmt.Fprintln(os.Stderr, "http error:", err)
		return
	}
	defer resp.Body.Close()

	rd := bufio.NewReader(resp.Body)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			return
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		raw := strings.TrimPrefix(line, "data: ")
		raw = strings.TrimSpace(raw)

		var rec telemetry.LogRecordV2
		if err := json.Unmarshal([]byte(raw), &rec); err != nil {
			continue
		}
		printLog(rec, *pretty, *compact)
	}
}

func urlQueryEsc(s string) string {
	r := strings.ReplaceAll(s, " ", "%20")
	r = strings.ReplaceAll(r, "#", "%23")
	return r
}

func printLog(r telemetry.LogRecordV2, pretty, compact bool) {
	ts := r.Time.Format(time.RFC3339)
	lv := r.Level.String()
	color := colorForLevel(lv)
	reset := "\033[0m"

	if compact {
		fmt.Printf("%s%s%-5s%s %s %s/%s %s\n",
			color, "", lv, reset, ts, r.Scope, r.KernelID, r.Message,
		)
		return
	}

	b, _ := json.Marshal(r.Fields)
	if pretty {
		var j any
		if json.Unmarshal(b, &j) == nil {
			b2, _ := json.MarshalIndent(j, "", "  ")
			b = b2
		}
	}
	fmt.Printf("%s%-5s%s %s %-8s %s/%s [%s] %s | %s\n",
		color, lv, reset, ts, r.Scope, r.KernelID, r.Component, r.Trace, r.Message, string(b),
	)
}

func colorForLevel(lv string) string {
	switch strings.ToUpper(lv) {
	case "DEBUG":
		return "\033[36m" // cyan
	case "INFO":
		return "\033[32m" // green
	case "WARN":
		return "\033[33m" // yellow
	case "ERROR":
		return "\033[31m" // red
	default:
		return "\033[0m"
	}
}

func cmdKernelsList(args []string) {
	fs := flag.NewFlagSet("kernels list", flag.ExitOnError)
	httpURL := fs.String("http", defaultHTTP(), "Base URL admin HTTP")
	_ = fs.Parse(args)

	url := strings.TrimRight(*httpURL, "/") + "/admin/kernels"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "http error:", err)
		return
	}
	defer resp.Body.Close()
	ioCopy(os.Stdout, resp.Body)
}

func cmdKernelsHealth(args []string) {
	fs := flag.NewFlagSet("kernels health", flag.ExitOnError)
	httpURL := fs.String("http", defaultHTTP(), "Base URL admin HTTP")
	_ = fs.Parse(args)

	url := strings.TrimRight(*httpURL, "/") + "/admin/health"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "http error:", err)
		return
	}
	defer resp.Body.Close()
	ioCopy(os.Stdout, resp.Body)
}

func cmdKernelsAction(args []string, action string) {
	fs := flag.NewFlagSet("kernels "+action, flag.ExitOnError)
	httpURL := fs.String("http", defaultHTTP(), "Base URL admin HTTP")
	id := fs.String("id", "", "Kernel ID")
	_ = fs.Parse(args)

	if *id == "" {
		fmt.Fprintln(os.Stderr, "--id is required")
		return
	}
	url := fmt.Sprintf("%s/admin/kernels/%s/%s", strings.TrimRight(*httpURL, "/"), *id, action)
	req, _ := http.NewRequest(http.MethodPost, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "http error:", err)
		return
	}
	defer resp.Body.Close()
	ioCopy(os.Stdout, resp.Body)
}

// маленькая утилита без лишних зависимостей
func ioCopy(dst *os.File, src io.Reader) {
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			dst.Write(buf[:n])
		}
		if err != nil {
			return
		}
	}
}
