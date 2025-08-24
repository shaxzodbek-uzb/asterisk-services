#!/bin/bash

# Run script for AMI-based webhook forwarder
# This captures ALL call events without needing dialplan changes!

echo "🚀 Starting AMI-based Asterisk Webhook Forwarder"
echo "==============================================="

# Check if config.env exists
if [ ! -f "config.env" ]; then
    echo "⚠️  Config file 'config.env' not found!"
    echo "📝 Creating config.env from template..."
    
    if [ -f "config.env.template" ]; then
        cp config.env.template config.env
        echo "✅ Created config.env from template"
        echo "📝 Please edit config.env with your actual credentials before running again."
        echo ""
        echo "Required settings:"
        echo "  - WEBHOOK_URL (your webhook endpoint)"
        echo "  - AMI_USER (from /etc/asterisk/manager.conf)"
        echo "  - AMI_PASS (from /etc/asterisk/manager.conf)"
        echo ""
        echo "Then run: ./run-ami-forwarder.sh"
        exit 1
    else
        echo "❌ No config template found. Please create config.env manually."
        exit 1
    fi
fi

echo "📁 Using configuration from config.env"
echo "🔨 Building AMI webhook forwarder..."

# Build the AMI forwarder
if ! go build -o ami-webhook-forwarder ami-webhook-forwarder.go; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✅ Build successful!"
echo "🚀 Starting forwarder..."
echo ""

# Run the AMI forwarder (it will load config.env automatically)
./ami-webhook-forwarder
