package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// loadDotEnv reads a local .env file (if present) and sets environment variables.
func loadDotEnv() {
	f, err := os.Open(".env")
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// Remove optional surrounding quotes
		if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
			val = strings.Trim(val, "\"")
		} else if strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'") {
			val = strings.Trim(val, "'")
		}
		os.Setenv(key, val)
	}
}

var logger *slog.Logger
var httpRequestCounter metric.Int64Counter
var httpRequestDurationHistogram metric.Float64Histogram

func initMetrics() error {
	// Create Prometheus exporter
	exporter, err := prometheus.New()
	if err != nil {
		return err
	}

	// Create metric provider with Prometheus exporter
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	otel.SetMeterProvider(meterProvider)

	// Create meter
	meter := meterProvider.Meter("hello-service")

	// Create counters and histograms
	httpRequestCounter, err = meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		return err
	}

	httpRequestDurationHistogram, err = meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
	)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// Load .env (if present) so environment variables are available.
	loadDotEnv()

	// Initialize OpenTelemetry metrics
	if err := initMetrics(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize metrics: %v\n", err)
		os.Exit(1)
	}

	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))
	logger.Info("Application started")

	// Wrap handlers with OpenTelemetry
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/health", handleHealth)
	mux.Handle("/metrics", promhttp.Handler())

	// Apply OpenTelemetry HTTP instrumentation
	handler := otelhttp.NewHandler(mux, "hello-service")

	// Get port from environment or default to 8050
	port := os.Getenv("PORT")
	if port == "" {
		port = "8050"
	}

	addr := ":" + port
	logger.Info("HTTP server listening", "port", port, "endpoints", []string{"/", "/health", "/metrics"})
	if err := http.ListenAndServe(addr, handler); err != nil {
		logger.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get environment variables
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "not set"
	}

	message := "Hello, World!"

	// Get server name
	serverName, _ := os.Hostname()

	// Get IP address
	ipAddr := getLocalIP()

	// Get current timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05 MST")

	// Read settings.json
	settingsHTML := ""
	file, err := os.Open("settings.json")
	if err == nil {
		defer file.Close()
		data, err := io.ReadAll(file)
		if err == nil {
			settings := make(map[string]interface{})
			if err := json.Unmarshal(data, &settings); err == nil {
				settingsHTML = "<h2>Json Settings</h2><div class='settings-box'>"

				// Sort keys for consistent order
				keys := make([]string, 0, len(settings))
				for k := range settings {
					keys = append(keys, k)
				}
				sort.Strings(keys)

				// Display settings in sorted order
				for _, k := range keys {
					settingsHTML += fmt.Sprintf("<div><span class='label'>%s:</span><span class='value'>%v</span></div>", k, settings[k])
				}
				settingsHTML += "</div>"
			}
		}
	}

	// Read input.txt
	inputHTML := ""
	inputFile, err := os.Open("input.txt")
	if err == nil {
		defer inputFile.Close()
		inputData, err := io.ReadAll(inputFile)
		if err == nil {
			inputHTML = "<h2>Text File Contents</h2><div class='settings-box'>"
			inputHTML += fmt.Sprintf("<div><pre style='margin:0; white-space: pre-wrap; word-wrap: break-word;'>%s</pre></div>", string(inputData))
			inputHTML += "</div>"
		}
	}

	// Fetch and display AWS Secrets
	secretsHTML := ""
	secretData, err := getAWSSecret("eso-secret-hello")
	if err == nil {
		secretsHTML = "<h2>AWS Secrets</h2><div class='settings-box'>"

		// Sort keys for consistent order
		secretKeys := make([]string, 0, len(secretData))
		for k := range secretData {
			secretKeys = append(secretKeys, k)
		}
		sort.Strings(secretKeys)

		// Display secrets in sorted order
		for _, k := range secretKeys {
			secretsHTML += fmt.Sprintf("<div><span class='label'>%s:</span><span class='value'>%v</span></div>", k, secretData[k])
		}
		secretsHTML += "</div>"
	}

	// Build HTML response
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<title>Hello Application</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 20px; }
		.container { max-width: 600px; margin: 0 auto; }
		.info-box { background: #f0f0f0; padding: 15px; border-radius: 5px; margin: 10px 0; }
		.settings-box { background: #f9f9f9; padding: 15px; border-radius: 5px; margin: 10px 0; border-left: 4px solid #007bff; }
		.label { font-weight: bold; color: #333; }
		.value { color: #666; margin-left: 10px; word-break: break-all; }
		h2 { margin-top: 20px; font-size: 1.2em; color: #333; }
		div { margin: 8px 0; }
	</style>
</head>
<body>
	<div class="container">
		<h1>%s</h1>
		<div class="info-box">
			<div><span class="label">Server Name:</span><span class="value">%s</span></div>
			<div><span class="label">IP Address:</span><span class="value">%s</span></div>
			<div><span class="label">Timestamp:</span><span class="value">%s</span></div>
		</div>
		%s
		%s
		%s
	</div>
</body>
</html>
`, message, environment, serverName, ipAddr, timestamp, inputHTML, settingsHTML, secretsHTML)

	fmt.Fprint(w, html)

	// Record metrics
	duration := time.Since(start).Seconds()
	httpRequestCounter.Add(ctx, 1)
	httpRequestDurationHistogram.Record(ctx, duration)

	logger.InfoContext(ctx, "Request completed",
		"path", r.URL.Path,
		"method", r.Method,
		"duration_seconds", duration)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"healthy"}`)

	// Record metrics
	duration := time.Since(start).Seconds()
	httpRequestCounter.Add(ctx, 1)
	httpRequestDurationHistogram.Record(ctx, duration)

	logger.InfoContext(ctx, "Health check",
		"path", r.URL.Path,
		"method", r.Method,
		"duration_seconds", duration)
}

func getLocalIP() string {
	// Try to connect to a public DNS to determine local IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "unknown"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
