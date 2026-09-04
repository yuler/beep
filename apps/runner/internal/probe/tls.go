package probe

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type TLSProbe struct{}

func NewTLSProbe() *TLSProbe {
	return &TLSProbe{}
}

func (p *TLSProbe) Run(ctx context.Context, config map[string]any) *Signal {
	targetHost, _ := config["host"].(string)
	if targetHost == "" {
		if targetURL, _ := config["target_url"].(string); targetURL != "" {
			if u, err := url.Parse(targetURL); err == nil {
				targetHost = u.Host
			}
		}
	}

	if targetHost == "" {
		return ErrorSignal("Missing TLS target host", "host config parameter is required", nil)
	}

	// Ensure port
	host := targetHost
	port := "443"
	if strings.Contains(targetHost, ":") {
		h, p, err := net.SplitHostPort(targetHost)
		if err == nil {
			host = h
			port = p
		}
	}
	addr := net.JoinHostPort(host, port)

	warnDays := 14
	if val, ok := config["days_remaining_threshold"]; ok {
		switch v := val.(type) {
		case float64:
			warnDays = int(v)
		case int:
			warnDays = v
		case string:
			if s, err := strconv.Atoi(v); err == nil {
				warnDays = s
			}
		}
	}

	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	}

	start := time.Now()
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	latencyMs := time.Since(start).Milliseconds()

	if err != nil {
		return AlertingSignal(
			fmt.Sprintf("TLS handshake failed for %s", host),
			err.Error(),
			map[string]any{
				"latency_ms": latencyMs,
			},
		)
	}
	defer conn.Close()

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return ErrorSignal("No peer certificates found", fmt.Sprintf("Server at %s returned zero certificates", addr), nil)
	}

	cert := state.PeerCertificates[0]
	now := time.Now()
	daysRemaining := int(cert.NotAfter.Sub(now).Hours() / 24)

	metrics := map[string]any{
		"days_remaining": daysRemaining,
		"latency_ms":     latencyMs,
		"issuer":         cert.Issuer.CommonName,
		"expires_at":     cert.NotAfter.UTC().Format(time.RFC3339),
	}

	if daysRemaining < 0 {
		return AlertingSignal(
			fmt.Sprintf("SSL/TLS certificate for %s EXPIRED", host),
			fmt.Sprintf("Certificate expired %d days ago on %s", -daysRemaining, cert.NotAfter.Format("2006-01-02")),
			metrics,
		)
	}

	if daysRemaining <= warnDays {
		return AlertingSignal(
			fmt.Sprintf("SSL/TLS certificate for %s expiring in %d days", host, daysRemaining),
			fmt.Sprintf("Certificate expires on %s (threshold: %d days)", cert.NotAfter.Format("2006-01-02"), warnDays),
			metrics,
		)
	}

	return OkSignal(
		fmt.Sprintf("SSL/TLS valid (%d days remaining)", daysRemaining),
		fmt.Sprintf("Certificate for %s is valid until %s", host, cert.NotAfter.Format("2006-01-02")),
		metrics,
	)
}
