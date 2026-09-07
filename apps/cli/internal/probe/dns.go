package probe

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

type DNSProbe struct{}

func NewDNSProbe() *DNSProbe {
	return &DNSProbe{}
}

func (p *DNSProbe) Run(ctx context.Context, config map[string]any) *Signal {
	hostname, _ := config["hostname"].(string)
	if hostname == "" {
		if h, _ := config["host"].(string); h != "" {
			hostname = h
		}
	}
	if hostname == "" {
		return ErrorSignal("Missing hostname", "hostname config parameter is required", nil)
	}

	resolver := net.DefaultResolver
	start := time.Now()
	ips, err := resolver.LookupHost(ctx, hostname)
	latencyMs := time.Since(start).Milliseconds()

	metrics := map[string]any{
		"latency_ms": latencyMs,
	}

	if err != nil {
		return AlertingSignal(
			fmt.Sprintf("DNS resolution failed for %s", hostname),
			err.Error(),
			metrics,
		)
	}

	metrics["resolved_ips"] = strings.Join(ips, ", ")

	return OkSignal(
		fmt.Sprintf("DNS %s resolved (%dms)", hostname, latencyMs),
		fmt.Sprintf("Resolved IPs: %s", strings.Join(ips, ", ")),
		metrics,
	)
}
