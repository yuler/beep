package probe

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"
)

type TCPProbe struct{}

func NewTCPProbe() *TCPProbe {
	return &TCPProbe{}
}

func (p *TCPProbe) Run(ctx context.Context, config map[string]any) *Signal {
	host, _ := config["host"].(string)
	if host == "" {
		return ErrorSignal("Missing TCP host", "host config parameter is required", nil)
	}

	port := 80
	if val, ok := config["port"]; ok {
		switch v := val.(type) {
		case float64:
			port = int(v)
		case int:
			port = v
		case string:
			if s, err := strconv.Atoi(v); err == nil {
				port = s
			}
		}
	}

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	dialer := &net.Dialer{Timeout: 10 * time.Second}

	start := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	latencyMs := time.Since(start).Milliseconds()

	metrics := map[string]any{
		"latency_ms": latencyMs,
		"port":       port,
	}

	if err != nil {
		return AlertingSignal(
			fmt.Sprintf("TCP connection to %s failed", addr),
			err.Error(),
			metrics,
		)
	}
	defer conn.Close()

	return OkSignal(
		fmt.Sprintf("TCP %s open (%dms)", addr, latencyMs),
		fmt.Sprintf("Successfully connected to TCP port %d on %s", port, host),
		metrics,
	)
}
