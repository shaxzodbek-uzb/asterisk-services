# Asterisk Webhook Forwarder

A lightweight Go application that connects to Asterisk's ARI (Asterisk REST Interface) to receive real-time events and forwards them to a webhook URL.

## Features

- Connects to Asterisk ARI via WebSocket for real-time events
- Forwards all Asterisk events to a configurable webhook URL
- Automatic ARI application registration
- Graceful shutdown handling
- Configurable via environment variables
- JSON payload formatting for webhooks

## Prerequisites

- Go 1.21 or higher
- Asterisk server with ARI enabled
- Access to Asterisk Manager credentials

## Configuration

The application is configured via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `ASTERISK_HOST` | `localhost` | Asterisk server hostname |
| `ASTERISK_PORT` | `8088` | Asterisk ARI port |
| `ASTERISK_USER` | `admin` | Asterisk manager username |
| `ASTERISK_PASS` | `admin` | Asterisk manager password |
| `ARI_APP_NAME` | `webhook-forwarder` | ARI application name |
| `WEBHOOK_URL` | `http://localhost:3000/webhook` | Target webhook URL |

## Installation & Usage

1. **Install dependencies:**
   ```bash
   go mod tidy
   ```

2. **Build the application:**
   ```bash
   go build -o asterisk-webhook-forwarder asterisk-webhook-forwarder.go
   ```

3. **Set environment variables:**
   ```bash
   export WEBHOOK_URL="https://your-server.com/asterisk-webhook"
   export ASTERISK_HOST="your-asterisk-server.com"
   export ASTERISK_USER="your-username"
   export ASTERISK_PASS="your-password"
   ```

4. **Run the application:**
   ```bash
   ./asterisk-webhook-forwarder
   ```

   Or use the provided example script:
   ```bash
   ./run-example.sh
   ```

## Webhook Payload Format

Events are sent to your webhook URL as POST requests with the following JSON structure:

```json
{
  "source": "asterisk-ari",
  "event_type": "StasisStart",
  "timestamp": "2023-12-01T10:30:00Z",
  "data": {
    "type": "StasisStart",
    "timestamp": "2023-12-01T10:30:00.123Z",
    "application": "webhook-forwarder",
    "channel": {
      "id": "1670837896.0",
      "name": "PJSIP/1001-00000000",
      "state": "Up",
      "caller": {
        "name": "John Doe",
        "number": "1001"
      }
    }
  }
}
```

## Common Event Types

The forwarder will capture and forward various Asterisk events including:

- `StasisStart` - Channel enters application
- `StasisEnd` - Channel leaves application  
- `ChannelStateChange` - Channel state changes
- `ChannelHangupRequest` - Hangup requested
- `ChannelDestroyed` - Channel destroyed
- `BridgeCreated` - Bridge created
- `BridgeDestroyed` - Bridge destroyed
- And many more...

## Error Handling

- Connection failures are logged and the application will exit
- Webhook delivery failures are logged but don't stop the application
- JSON parsing errors are logged and the problematic event is skipped
- Graceful shutdown on SIGINT/SIGTERM

## Development

To modify the application:

1. Edit `asterisk-webhook-forwarder.go`
2. Test with `go run asterisk-webhook-forwarder.go`
3. Build with `go build`

## License

This project is provided as-is for educational and development purposes.
