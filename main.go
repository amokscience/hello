package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
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
		environment = "development"
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
		.label { font-weight: bold; color: #333; }
		.value { color: #666; margin-left: 10px; }
	</style>
</head>
<body>
	<div class="container">
		<h1>%s</h1>
		<div class="info-box">
			<div><span class="label">Environment:</span><span class="value">%s</span></div>
			<div><span class="label">Server Name:</span><span class="value">%s</span></div>
			<div><span class="label">IP Address:</span><span class="value">%s</span></div>
		</div>
	</div>
</body>
</html>
`, message, environment, serverName, ipAddr)

	fmt.Fprint(w, html)
	logger.Info("Request handled", "path", r.URL.Path, "method", r.Method)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"healthy"}`)
	logger.Info("Health check", "path", r.URL.Path)
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
