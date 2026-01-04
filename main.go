package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sort"
	"time"
)

var logger *slog.Logger

func main() {
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Application started")

	// Set up HTTP routes
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/health", handleHealth)

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	logger.Info("HTTP server listening", "port", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get environment variables
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "not set"
	}

	customMessage := os.Getenv("HELLO")
	if customMessage == "" {
		customMessage = "Hello, World!"
	}

	// Append environment to message
	message := fmt.Sprintf("%s - %s", customMessage, environment)

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
				settingsHTML = "<h2>Settings</h2><div class='settings-box'>"

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
			<div><span class="label">Environment:</span><span class="value">%s</span></div>
			<div><span class="label">Server Name:</span><span class="value">%s</span></div>
			<div><span class="label">IP Address:</span><span class="value">%s</span></div>
			<div><span class="label">Timestamp:</span><span class="value">%s</span></div>
		</div>
		%s
		%s
	</div>
</body>
</html>
`, message, environment, serverName, ipAddr, timestamp, settingsHTML, secretsHTML)

	fmt.Fprint(w, html)
	logger.Info("Request handled", "path", r.URL.Path, "method", r.Method)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"healthy"}`)
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
