package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

type Config struct {
	AsteriskHost string
	AsteriskPort string
	AsteriskUser string
	AsteriskPass string
	AppName      string
	WebhookURL   string
}

type ARIEvent struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Event     interface{} `json:",omitempty"`
	Data      interface{} `json:",omitempty"`
}

type WebhookPayload struct {
	Source    string      `json:"source"`
	EventType string      `json:"event_type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

func loadConfig() *Config {
	config := &Config{
		AsteriskHost: getEnv("ASTERISK_HOST", "localhost"),
		AsteriskPort: getEnv("ASTERISK_PORT", "8088"),
		AsteriskUser: getEnv("ASTERISK_USER", "admin"),
		AsteriskPass: getEnv("ASTERISK_PASS", "admin"),
		AppName:      getEnv("ARI_APP_NAME", "webhook-forwarder"),
		WebhookURL:   getEnv("WEBHOOK_URL", "http://localhost:3000/webhook"),
	}

	if config.WebhookURL == "" {
		log.Fatal("WEBHOOK_URL environment variable is required")
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func connectToARI(config *Config) (*websocket.Conn, error) {
	// Build WebSocket URL for ARI events
	u := url.URL{
		Scheme: "ws",
		Host:   config.AsteriskHost + ":" + config.AsteriskPort,
		Path:   "/ari/events",
		RawQuery: fmt.Sprintf("app=%s&api_key=%s:%s", 
			config.AppName, config.AsteriskUser, config.AsteriskPass),
	}

	log.Printf("Connecting to ARI WebSocket: %s", u.String())

	// Set up WebSocket connection
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ARI WebSocket: %w", err)
	}

	log.Println("Successfully connected to Asterisk ARI")
	return conn, nil
}

func sendToWebhook(config *Config, payload WebhookPayload) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequest("POST", config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Asterisk-Webhook-Forwarder/1.0")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status: %d", resp.StatusCode)
	}

	return nil
}

func handleEvents(conn *websocket.Conn, config *Config) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading WebSocket message: %v", err)
			return
		}

		// Parse the ARI event
		var event map[string]interface{}
		if err := json.Unmarshal(message, &event); err != nil {
			log.Printf("Error parsing ARI event JSON: %v", err)
			continue
		}

		// Extract event type
		eventType, ok := event["type"].(string)
		if !ok {
			log.Println("Event missing type field, skipping")
			continue
		}

		// Create webhook payload
		payload := WebhookPayload{
			Source:    "asterisk-ari",
			EventType: eventType,
			Timestamp: time.Now().UTC(),
			Data:      event,
		}

		// Send to webhook
		if err := sendToWebhook(config, payload); err != nil {
			log.Printf("Failed to send event '%s' to webhook: %v", eventType, err)
		} else {
			log.Printf("Successfully forwarded event '%s' to webhook", eventType)
		}
	}
}

func registerApplication(config *Config) error {
	// Register the ARI application
	apiURL := fmt.Sprintf("http://%s:%s/ari/applications/%s", 
		config.AsteriskHost, config.AsteriskPort, config.AppName)

	req, err := http.NewRequest("PUT", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create application registration request: %w", err)
	}

	req.SetBasicAuth(config.AsteriskUser, config.AsteriskPass)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to register ARI application: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to register ARI application, status: %d", resp.StatusCode)
	}

	log.Printf("Successfully registered ARI application: %s", config.AppName)
	return nil
}

func main() {
	log.Println("Starting Asterisk Webhook Forwarder...")

	// Load configuration
	config := loadConfig()
	
	log.Printf("Configuration loaded:")
	log.Printf("  Asterisk: %s:%s", config.AsteriskHost, config.AsteriskPort)
	log.Printf("  ARI App: %s", config.AppName)
	log.Printf("  Webhook URL: %s", config.WebhookURL)

	// Register ARI application
	if err := registerApplication(config); err != nil {
		log.Fatalf("Failed to register ARI application: %v", err)
	}

	// Connect to ARI WebSocket
	conn, err := connectToARI(config)
	if err != nil {
		log.Fatalf("Failed to connect to ARI: %v", err)
	}
	defer conn.Close()

	// Set up graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Start event handling in a goroutine
	go handleEvents(conn, config)

	log.Println("Webhook forwarder is running. Press Ctrl+C to stop...")

	// Wait for interrupt signal
	<-interrupt
	log.Println("Shutting down gracefully...")

	// Close WebSocket connection
	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	time.Sleep(time.Second)

	log.Println("Shutdown complete")
}
