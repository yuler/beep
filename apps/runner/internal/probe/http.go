package probe

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HTTPProbe struct {
	client *http.Client
}

func NewHTTPProbe() *HTTPProbe {
	return &HTTPProbe{
		client: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return fmt.Errorf("stopped after 5 redirects")
				}
				return nil
			},
		},
	}
}

func (p *HTTPProbe) Run(ctx context.Context, config map[string]any) *Signal {
	targetURL, _ := config["target_url"].(string)
	if targetURL == "" {
		return ErrorSignal("Missing target URL", "target_url config parameter is required", nil)
	}

	method, _ := config["method"].(string)
	if method == "" {
		method = "GET"
	}
	method = strings.ToUpper(method)

	expectedStatus := 200
	if val, ok := config["expected_status"]; ok {
		switch v := val.(type) {
		case float64:
			expectedStatus = int(v)
		case int:
			expectedStatus = v
		case string:
			if s, err := strconv.Atoi(v); err == nil {
				expectedStatus = s
			}
		}
	}

	timeoutSec := 10
	if val, ok := config["timeout_seconds"]; ok {
		if v, ok := val.(float64); ok && v > 0 {
			timeoutSec = int(v)
		}
	}

	probeCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(probeCtx, method, targetURL, nil)
	if err != nil {
		return ErrorSignal("Invalid HTTP request", err.Error(), nil)
	}
	req.Header.Set("User-Agent", "Beep-Runner/1.0")

	start := time.Now()
	resp, err := p.client.Do(req)
	latency := time.Since(start)
	latencyMs := latency.Milliseconds()

	if err != nil {
		return AlertingSignal(
			fmt.Sprintf("%s is unreachable", targetURL),
			err.Error(),
			map[string]any{
				"latency_ms": latencyMs,
				"status":     0,
			},
		)
	}
	defer resp.Body.Close()

	// Read up to 8KB of body
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	bodyStr := string(bodyBytes)

	metrics := map[string]any{
		"latency_ms":  latencyMs,
		"status_code": resp.StatusCode,
	}

	if resp.StatusCode != expectedStatus {
		return AlertingSignal(
			fmt.Sprintf("HTTP %d (expected %d)", resp.StatusCode, expectedStatus),
			fmt.Sprintf("Expected status %d, got %d from %s", expectedStatus, resp.StatusCode, targetURL),
			metrics,
		)
	}

	// Match keyword if configured
	if keyword, _ := config["body_contains"].(string); keyword != "" {
		if !strings.Contains(bodyStr, keyword) {
			return AlertingSignal(
				"Body content mismatch",
				fmt.Sprintf("Response body does not contain expected keyword %q", keyword),
				metrics,
			)
		}
	}

	return OkSignal(
		fmt.Sprintf("HTTP %d OK (%dms)", resp.StatusCode, latencyMs),
		fmt.Sprintf("Successfully connected to %s", targetURL),
		metrics,
	)
}
