# Asterisk AMI Webhook Forwarder

A lightweight Go application that connects to Asterisk's AMI (Asterisk Manager Interface) to receive real-time call events and forwards them to a webhook URL.

## Features

- Connects to Asterisk AMI for real-time call events
- Captures ALL call events without requiring dialplan changes
- Forwards call events to a configurable webhook URL
- Automatic reconnection on connection loss
- Graceful shutdown handling
- Configurable via environment variables
- JSON payload formatting for webhooks
- Heartbeat monitoring

## Prerequisites

- Go 1.21 or higher
- Asterisk server with AMI enabled
- AMI user credentials configured in manager.conf

## Configuration

The application can be configured in two ways:

### 1. Configuration File (Recommended)
Create a `config.env` file with your settings:

```bash
# Use the setup script for guided configuration
./setup.sh
```

### 2. Environment Variables
Set environment variables directly:

| Variable | Default | Description |
|----------|---------|-------------|
| `ASTERISK_HOST` | `localhost` | Asterisk server hostname |
| `AMI_PORT` | `5038` | Asterisk AMI port |
| `AMI_USER` | `admin` | AMI username from manager.conf |
| `AMI_PASS` | `admin` | AMI password from manager.conf |
| `WEBHOOK_URL` | *required* | Target webhook URL |
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `AMI_TIMEOUT` | `60` | AMI connection timeout in seconds |
| `WEBHOOK_TIMEOUT` | `10` | Webhook request timeout in seconds |

## AMI Setup

1. **Configure AMI in `/etc/asterisk/manager.conf`:**
   ```ini
   [general]
   enabled = yes
   port = 5038
   bindaddr = 0.0.0.0
   allowmultiple = yes

   [webhook-user]  ; Create a dedicated user
   secret = your-secure-password
   read = all
   write = all
   ```

2. **Reload Asterisk configuration:**
   ```bash
   sudo asterisk -rx "manager reload"
   ```

## Quick Start

### 1. Easy Setup (Recommended)
```bash
# Make scripts executable
chmod +x *.sh

# Run guided setup
./setup.sh

# Start the forwarder
./run-ami-forwarder.sh
```

### 2. Manual Setup

1. **Create configuration file:**
   ```bash
   cp config.env.template config.env
   # Edit config.env with your settings
   ```

2. **Build and run:**
   ```bash
   go build -o ami-webhook-forwarder ami-webhook-forwarder.go
   ./ami-webhook-forwarder
   ```

### 3. Using Environment Variables Only
   ```bash
   export WEBHOOK_URL="https://your-server.com/asterisk-webhook"
   export ASTERISK_HOST="your-asterisk-server.com"
   export AMI_USER="webhook-user"
   export AMI_PASS="your-secure-password"
   
   go run ami-webhook-forwarder.go
   ```

## Webhook Payload Format

Events are sent to your webhook URL as POST requests with the following JSON structure:

```json
{
  "source": "asterisk-ami",
  "event_type": "Newchannel",
  "timestamp": "2023-12-01T10:30:00Z",
  "data": {
    "Event": "Newchannel",
    "Privilege": "call,all",
    "Channel": "PJSIP/1001-00000001",
    "ChannelState": "0",
    "ChannelStateDesc": "Down",
    "CallerIDNum": "1001",
    "CallerIDName": "John Doe",
    "ConnectedLineNum": "",
    "ConnectedLineName": "",
    "Uniqueid": "1670837896.0",
    "Context": "from-internal",
    "Exten": "1002"
  }
}
```

## Captured Event Types

The AMI forwarder captures and forwards all call-related events including:

- `Newchannel` - New call channel created
- `Hangup` - Call hangup/termination
- `DialBegin` - Outbound call started
- `DialEnd` - Outbound call finished  
- `Bridge` / `Unbridge` - Calls connected/disconnected
- `NewCallerid` - Caller ID changes
- `NewState` - Channel state changes
- `Transfer` / `AttendedTransfer` / `BlindTransfer` - Call transfers
- `Hold` / `Unhold` - Call hold events
- `DTMF` - Keypad input
- `VoicemailUserEntry` - Voicemail access
- `CDR` / `CEL` - Call detail records
- And many more call-related events...

## Security

- **Credentials:** Store sensitive credentials in `config.env` (not tracked by git)
- **File permissions:** Set restrictive permissions on config file:
  ```bash
  chmod 600 config.env
  ```
- **AMI user:** Create dedicated AMI user with minimal required permissions
- **Network:** Use firewall rules to restrict AMI access (port 5038)
- **HTTPS:** Always use HTTPS for webhook URLs in production

## Error Handling

- Connection failures are logged and the application will exit
- Webhook delivery failures are logged but don't stop the application
- JSON parsing errors are logged and the problematic event is skipped
- Graceful shutdown on SIGINT/SIGTERM
- Automatic reconnection on connection loss

## Testing

```bash
# Test configuration and AMI connection
./setup.sh

# Generate test events
./test-events.sh

# Check logs for successful webhook deliveries
```

## Troubleshooting

### Common Issues

1. **AMI Connection Failed**
   - Check if Asterisk is running: `sudo asterisk -rx "core show version"`
   - Verify AMI is enabled in `/etc/asterisk/manager.conf`
   - Test port connectivity: `telnet localhost 5038`

2. **Authentication Failed**
   - Verify AMI credentials in `config.env`
   - Check AMI user permissions in `/etc/asterisk/manager.conf`
   - Reload manager: `sudo asterisk -rx "manager reload"`

3. **No Events Received**
   - Make test calls to generate events
   - Check event filtering in `shouldProcessEvent()` function
   - Verify webhook URL is reachable

## Development

To modify the application:

1. Edit `ami-webhook-forwarder.go`
2. Test with `go run ami-webhook-forwarder.go`
3. Build with `go build`

## License

This project is provided as-is for educational and development purposes.
