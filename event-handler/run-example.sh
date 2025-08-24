#!/bin/bash

# Example script to run the Asterisk Webhook Forwarder
# Make sure to adjust these environment variables for your setup

export ASTERISK_HOST="localhost"
export ASTERISK_PORT="8088"
export ASTERISK_USER="admin"
export ASTERISK_PASS="admin"
export ARI_APP_NAME="webhook-forwarder"
export WEBHOOK_URL="http://localhost:3000/asterisk-events"

echo "Starting Asterisk Webhook Forwarder with the following configuration:"
echo "  Asterisk Host: $ASTERISK_HOST:$ASTERISK_PORT"
echo "  ARI App Name: $ARI_APP_NAME"
echo "  Webhook URL: $WEBHOOK_URL"
echo ""

# Build and run the application
go build -o asterisk-webhook-forwarder asterisk-webhook-forwarder.go
./asterisk-webhook-forwarder
