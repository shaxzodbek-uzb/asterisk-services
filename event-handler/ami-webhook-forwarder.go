package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type AMIConfig struct {
	AsteriskHost string
	AsteriskPort string
	Username     string
	Password     string
	WebhookURL   string
}

type WebhookPayload struct {
	Source    string                 `json:"source"`
	EventType string                 `json:"event_type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

func loadAMIConfig() *AMIConfig {
	// Load from config file if exists, otherwise use environment variables
	loadConfigFile("config.env")

	config := &AMIConfig{
		AsteriskHost: getEnv("ASTERISK_HOST", "localhost"),
		AsteriskPort: getEnv("AMI_PORT", "5038"),
		Username:     getEnv("AMI_USER", "admin"),
		Password:     getEnv("AMI_PASS", "admin"),
		WebhookURL:   getEnv("WEBHOOK_URL", ""),
	}

	if config.WebhookURL == "" {
		log.Fatal("WEBHOOK_URL is required. Set it in config.env file or environment variable.")
	}

	log.Printf("Configuration loaded from config file and environment:")
	log.Printf("  Asterisk: %s:%s", config.AsteriskHost, config.AsteriskPort)
	log.Printf("  AMI User: %s", config.Username)
	log.Printf("  Webhook URL: %s", config.WebhookURL)

	return config
}

func loadConfigFile(filename string) {
	// Check if config file exists
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Config file '%s' not found, using environment variables only", filename)
		return
	}
	defer file.Close()

	log.Printf("Loading configuration from %s", filename)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			
			// Only set if environment variable is not already set
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading config file: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func connectToAMI(config *AMIConfig) (net.Conn, error) {
	// Connect to AMI
	conn, err := net.Dial("tcp", config.AsteriskHost+":"+config.AsteriskPort)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AMI: %w", err)
	}

	// Read welcome message
	reader := bufio.NewReader(conn)
	welcome, _ := reader.ReadString('\n')
	log.Printf("AMI Welcome: %s", strings.TrimSpace(welcome))

	// Send login
	loginMsg := fmt.Sprintf("Action: Login\r\nUsername: %s\r\nSecret: %s\r\n\r\n", 
		config.Username, config.Password)
	
	if _, err := conn.Write([]byte(loginMsg)); err != nil {
		return nil, fmt.Errorf("failed to send login: %w", err)
	}

	// Read login response
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read login response: %w", err)
		}
		line = strings.TrimSpace(line)
		
		if strings.Contains(line, "Message: Authentication accepted") {
			log.Println("Successfully authenticated with AMI")
			break
		}
		if strings.Contains(line, "Message: Authentication failed") {
			return nil, fmt.Errorf("AMI authentication failed")
		}
		if line == "" {
			break
		}
	}

	return conn, nil
}

func sendToWebhook(config *AMIConfig, payload WebhookPayload) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequest("POST", config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Asterisk-AMI-Webhook-Forwarder/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
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

func parseAMIEvent(eventText string) (string, map[string]interface{}) {
	lines := strings.Split(eventText, "\r\n")
	eventData := make(map[string]interface{})
	var eventType string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			
			if key == "Event" {
				eventType = value
			}
			eventData[key] = value
		}
	}

	return eventType, eventData
}

func handleAMIEvents(conn net.Conn, config *AMIConfig) {
	log.Println("Starting AMI event handler - listening for ALL Asterisk events...")
	reader := bufio.NewReader(conn)
	eventCount := 0
	lastHeartbeat := time.Now()
	
	var currentEvent strings.Builder
	
	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading AMI message: %v", err)
			// Try to reconnect after error
			return
		}

		// Reset read deadline
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		line = strings.TrimSpace(line)
		
		// Heartbeat every 60 seconds
		if time.Since(lastHeartbeat) > 60*time.Second {
			log.Printf("AMI Heartbeat: Connection alive, processed %d events", eventCount)
			lastHeartbeat = time.Now()
		}
		
		// Empty line indicates end of event
		if line == "" {
			eventText := currentEvent.String()
			if eventText != "" && strings.Contains(eventText, "Event:") {
				// Parse the event
				eventType, eventData := parseAMIEvent(eventText)
				
				// Only process call-related events
				if shouldProcessEvent(eventType) {
					eventCount++
					log.Printf("📞 Received AMI event #%d: %s", eventCount, eventType)
					
					// Create webhook payload
					payload := WebhookPayload{
						Source:    "asterisk-ami",
						EventType: eventType,
						Timestamp: time.Now().UTC(),
						Data:      eventData,
					}

					// Send to webhook
					if err := sendToWebhook(config, payload); err != nil {
						log.Printf("❌ Failed to send event '%s' to webhook: %v", eventType, err)
					} else {
						log.Printf("✅ Successfully forwarded event '%s' to webhook", eventType)
					}
				}
			}
			
			// Reset for next event
			currentEvent.Reset()
			continue
		}

		// Add line to current event
		currentEvent.WriteString(line + "\r\n")
	}
}

func shouldProcessEvent(eventType string) bool {
	// Process all call-related events
	callEvents := []string{
		"Newchannel", "Hangup", "DialBegin", "DialEnd", "Bridge", "Unbridge",
		"NewCallerid", "NewAccountCode", "NewExten", "NewState", "Dial",
		"AgentCalled", "AgentConnect", "QueueMemberAdded", "QueueMemberRemoved",
		"Hold", "Unhold", "MusicOnHoldStart", "MusicOnHoldStop", "Transfer",
		"AttendedTransfer", "BlindTransfer", "DTMF", "VoicemailUserEntry",
		"CEL", "CDR", "LocalBridge", "LocalOptimizationBegin", "LocalOptimizationEnd",
		"OriginateResponse", "ChannelTalkingStart", "ChannelTalkingStop",
		"BridgeCreate", "BridgeDestroy", "BridgeEnter", "BridgeLeave",
		"VarSet", "UserEvent", "Registry", "PeerStatus", "ContactStatus",
	}
	
	for _, event := range callEvents {
		if eventType == event {
			return true
		}
	}
	
	return false
}

func main() {
	log.Println("Starting Asterisk AMI Webhook Forwarder...")
	log.Println("This will capture ALL call events without needing dialplan changes!")

	// Load configuration
	config := loadAMIConfig()
	
	log.Printf("Configuration loaded:")
	log.Printf("  Asterisk AMI: %s:%s", config.AsteriskHost, config.AsteriskPort)
	log.Printf("  Username: %s", config.Username)
	log.Printf("  Webhook URL: %s", config.WebhookURL)

	// Connect to AMI
	conn, err := connectToAMI(config)
	if err != nil {
		log.Fatalf("Failed to connect to AMI: %v", err)
	}
	defer conn.Close()

	// Set up graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Start event handling in a goroutine
	go handleAMIEvents(conn, config)

	log.Println("AMI Webhook forwarder is running. Press Ctrl+C to stop...")

	// Wait for interrupt signal
	<-interrupt
	log.Println("Shutting down gracefully...")

	log.Println("Shutdown complete")
}
