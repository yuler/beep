package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"beep/internal/cmdutil"
	"beep/internal/config"
	"beep/internal/version"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// NewCmdAPI creates the 'beep api' command.
func NewCmdAPI() *cobra.Command {
	var (
		flagMethod    string
		flagFields    []string
		flagRawFields []string
		flagHeaders   []string
		flagInput     string
		flagInclude   bool
	)

	cmd := &cobra.Command{
		Use:   "api <path>",
		Short: "Make an authenticated HTTP request to the Beep API",
		Long: fmt.Sprintf(`Makes an authenticated HTTP request to the Beep API and prints the response.

The <path> argument can be a relative path (e.g. /api/v1/beeps or api/v1/beeps)
or a full URL. Relative paths are resolved against the configured server URL.

HTTP method inference:
  - If -X / --method is specified, that method is used.
  - If not specified, defaults to POST if -f, --raw-field, or --input is provided.
  - Otherwise defaults to GET.

Parameters and fields:
  - For GET requests, -f / --field key=value adds query parameters.
  - For other methods, -f key=value adds string fields to the JSON body,
    and --raw-field key=value parses the value as JSON (numbers, booleans, arrays, objects).
  - Use --input <file> (or --input - for stdin) to pass raw request body.

Examples:
  # List all beeps (GET /api/v1/beeps)
  %s api /api/v1/beeps

  # Filter with query params (GET /api/v1/beeps?status=active)
  %s api /api/v1/beeps -X GET -f status=active

  # Create a beep via POST
  %s api /api/v1/beeps \
    -f title="Standup reminder" \
    -f run_at="2026-10-01T09:00:00Z" \
    --raw-field metadata='{"category":"standup"}'

  # Include response HTTP headers
  %s api /api/v1/beeps -i`,
			config.BinaryName(), config.BinaryName(), config.BinaryName(), config.BinaryName()),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pathArg := strings.TrimSpace(args[0])
			if pathArg == "" {
				return fmt.Errorf("path cannot be empty")
			}

			cfg, err := cmdutil.LoadConfig(cmd)
			if err != nil {
				return err
			}

			// If AccessToken is empty but RunnerToken is present, reuse it for auth
			if cfg.AccessToken == "" && cfg.RunnerToken != "" {
				cfg.AccessToken = cfg.RunnerToken
			}

			targetURL := pathArg
			if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
				serverURL := strings.TrimRight(cfg.ServerURL, "/")
				if serverURL == "" {
					return fmt.Errorf("server URL is not configured. Please run '%s auth login' or specify --server", config.BinaryName())
				}
				if !strings.HasPrefix(targetURL, "/") {
					targetURL = "/" + targetURL
				}
				targetURL = serverURL + targetURL
			}

			// Determine HTTP method
			method := strings.ToUpper(strings.TrimSpace(flagMethod))
			if method == "" {
				if len(flagFields) > 0 || len(flagRawFields) > 0 || flagInput != "" {
					method = http.MethodPost
				} else {
					method = http.MethodGet
				}
			}

			var reqBody io.Reader
			contentType := "application/json"

			if method == http.MethodGet {
				if flagInput != "" {
					return fmt.Errorf("cannot use --input with GET request")
				}
				if len(flagFields) > 0 || len(flagRawFields) > 0 {
					u, err := url.Parse(targetURL)
					if err != nil {
						return fmt.Errorf("invalid URL %q: %w", targetURL, err)
					}
					q := u.Query()
					for _, item := range flagFields {
						k, v, ok := strings.Cut(item, "=")
						if !ok {
							return fmt.Errorf("invalid field parameter %q: must be KEY=VALUE", item)
						}
						q.Add(k, v)
					}
					for _, item := range flagRawFields {
						k, v, ok := strings.Cut(item, "=")
						if !ok {
							return fmt.Errorf("invalid raw-field parameter %q: must be KEY=VALUE", item)
						}
						q.Add(k, v)
					}
					u.RawQuery = q.Encode()
					targetURL = u.String()
				}
			} else {
				if flagInput != "" {
					if len(flagFields) > 0 || len(flagRawFields) > 0 {
						return fmt.Errorf("cannot combine --input with -f / --field or --raw-field")
					}
					var inputData []byte
					if flagInput == "-" {
						inputData, err = io.ReadAll(cmd.InOrStdin())
						if err != nil {
							return fmt.Errorf("failed to read from stdin: %w", err)
						}
					} else {
						inputData, err = os.ReadFile(flagInput)
						if err != nil {
							return fmt.Errorf("failed to read input file %s: %w", flagInput, err)
						}
					}
					reqBody = bytes.NewReader(inputData)
				} else if len(flagFields) > 0 || len(flagRawFields) > 0 {
					bodyMap := make(map[string]any)
					for _, item := range flagFields {
						k, v, ok := strings.Cut(item, "=")
						if !ok {
							return fmt.Errorf("invalid field parameter %q: must be KEY=VALUE", item)
						}
						bodyMap[k] = v
					}
					for _, item := range flagRawFields {
						k, v, ok := strings.Cut(item, "=")
						if !ok {
							return fmt.Errorf("invalid raw-field parameter %q: must be KEY=VALUE", item)
						}
						var jsonVal any
						if err := json.Unmarshal([]byte(v), &jsonVal); err == nil {
							bodyMap[k] = jsonVal
						} else {
							bodyMap[k] = v
						}
					}
					bodyBytes, err := json.Marshal(bodyMap)
					if err != nil {
						return fmt.Errorf("failed to encode request body: %w", err)
					}
					reqBody = bytes.NewReader(bodyBytes)
				}
			}

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, method, targetURL, reqBody)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			req.Header.Set("Accept", "application/json")
			req.Header.Set("User-Agent", fmt.Sprintf("Beep-CLI/%s (%s; %s)", version.Version, runtime.GOOS, runtime.GOARCH))
			if reqBody != nil {
				req.Header.Set("Content-Type", contentType)
			}

			// Apply custom headers first so we know what's explicitly set
			hasAuthHeader := false
			hasAccountHeader := false
			for _, h := range flagHeaders {
				k, v, ok := strings.Cut(h, ":")
				if !ok {
					return fmt.Errorf("invalid header %q: must be KEY:VALUE", h)
				}
				k = strings.TrimSpace(k)
				v = strings.TrimSpace(v)
				if strings.EqualFold(k, "authorization") {
					hasAuthHeader = true
				}
				if strings.EqualFold(k, "x-account-slug") {
					hasAccountHeader = true
				}
				req.Header.Set(k, v)
			}

			// Default Auth and Account headers if not overridden
			if !hasAuthHeader && cfg.AccessToken != "" {
				req.Header.Set("Authorization", "Bearer "+cfg.AccessToken)
			}
			if !hasAccountHeader && cfg.AccountSlug != "" {
				req.Header.Set("X-Account-Slug", cfg.AccountSlug)
			}

			httpClient := &http.Client{
				Timeout: 60 * time.Second,
			}

			resp, err := httpClient.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %w", err)
			}

			if flagInclude {
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s\r\n", resp.Proto, resp.Status)
				for k, values := range resp.Header {
					for _, v := range values {
						fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\r\n", k, v)
					}
				}
				fmt.Fprint(cmd.OutOrStdout(), "\r\n")
			}

			isTTY := isatty.IsTerminal(os.Stdout.Fd())
			output := respBody
			if isTTY && len(respBody) > 0 {
				var prettyJSON bytes.Buffer
				if err := json.Indent(&prettyJSON, respBody, "", "  "); err == nil {
					output = prettyJSON.Bytes()
				}
			}

			if len(output) > 0 {
				_, _ = cmd.OutOrStdout().Write(output)
				if !bytes.HasSuffix(output, []byte("\n")) {
					_, _ = cmd.OutOrStdout().Write([]byte("\n"))
				}
			}

			if resp.StatusCode >= 400 {
				return fmt.Errorf("HTTP %d (%s)", resp.StatusCode, resp.Status)
			}

			return nil
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			flagMethod = ""
			flagFields = nil
			flagRawFields = nil
			flagHeaders = nil
			flagInput = ""
			flagInclude = false
		},
	}

	cmd.Flags().StringVarP(&flagMethod, "method", "X", "", "HTTP method (e.g. GET, POST, PUT, DELETE)")
	cmd.Flags().StringArrayVarP(&flagFields, "field", "f", nil, "Add a string parameter in KEY=VALUE format")
	cmd.Flags().StringArrayVar(&flagRawFields, "raw-field", nil, "Add a typed/JSON parameter in KEY=VALUE format")
	cmd.Flags().StringArrayVarP(&flagHeaders, "header", "H", nil, "Add a custom HTTP request header in KEY:VALUE format")
	cmd.Flags().StringVar(&flagInput, "input", "", "The file to use as body for the HTTP request (use \"-\" to read from standard input)")
	cmd.Flags().BoolVarP(&flagInclude, "include", "i", false, "Include HTTP response status and headers in the output")

	return cmd
}
