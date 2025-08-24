#!/bin/bash

# Setup script for Asterisk AMI Webhook Forwarder
# This script helps configure your credentials and test the setup

echo "🛠️  Asterisk AMI Webhook Forwarder Setup"
echo "========================================"

# Step 1: Create config file
echo "📝 Step 1: Configuration File Setup"
echo ""

if [ -f "config.env" ]; then
    echo "⚠️  config.env already exists!"
    read -p "Do you want to overwrite it? (y/N): " overwrite
    if [[ ! "$overwrite" =~ ^[Yy]$ ]]; then
        echo "Using existing config.env"
    else
        rm config.env
    fi
fi

if [ ! -f "config.env" ]; then
    if [ -f "config.env.template" ]; then
        cp config.env.template config.env
        echo "✅ Created config.env from template"
    else
        echo "❌ Template not found, creating basic config..."
        cat > config.env << 'EOF'
# Asterisk AMI Webhook Forwarder Configuration
ASTERISK_HOST=localhost
AMI_PORT=5038
AMI_USER=admin
AMI_PASS=admin
WEBHOOK_URL=https://webhook-test.com/your-webhook-id
LOG_LEVEL=info
EOF
        echo "✅ Created basic config.env"
    fi
fi

echo ""

# Step 2: Interactive configuration
echo "📝 Step 2: Configure Your Settings"
echo ""

# Read current values
source config.env 2>/dev/null || true

# Function to update config value
update_config() {
    local key="$1"
    local current_value="$2"
    local prompt="$3"
    local new_value
    
    echo -n "$prompt"
    if [ -n "$current_value" ]; then
        echo -n " (current: $current_value)"
    fi
    echo -n ": "
    read new_value
    
    if [ -n "$new_value" ]; then
        # Update the config file
        if grep -q "^$key=" config.env; then
            # Key exists, update it
            sed -i.bak "s/^$key=.*/$key=$new_value/" config.env && rm config.env.bak
        else
            # Key doesn't exist, add it
            echo "$key=$new_value" >> config.env
        fi
        echo "✅ Updated $key"
    fi
}

echo "Configure your AMI settings:"
update_config "ASTERISK_HOST" "$ASTERISK_HOST" "Asterisk server hostname"
update_config "AMI_PORT" "$AMI_PORT" "AMI port"
update_config "AMI_USER" "$AMI_USER" "AMI username"
update_config "AMI_PASS" "$AMI_PASS" "AMI password"

echo ""
echo "Configure your webhook:"
update_config "WEBHOOK_URL" "$WEBHOOK_URL" "Webhook URL (REQUIRED)"

echo ""
echo "✅ Configuration complete!"
echo ""

# Step 3: Test AMI connection
echo "🧪 Step 3: Test AMI Connection"
echo ""

# Re-source the updated config
source config.env

if [ -z "$WEBHOOK_URL" ] || [[ "$WEBHOOK_URL" == *"your-webhook"* ]]; then
    echo "⚠️  WARNING: Please set a valid WEBHOOK_URL in config.env"
    echo "   You can use https://webhook.site to get a test URL"
fi

echo "Testing AMI connection to $ASTERISK_HOST:$AMI_PORT..."

# Try to connect to AMI port
if timeout 5 bash -c "</dev/tcp/$ASTERISK_HOST/$AMI_PORT" 2>/dev/null; then
    echo "✅ AMI port $AMI_PORT is accessible on $ASTERISK_HOST"
    
    # Test AMI credentials if possible
    echo "🔐 Testing AMI credentials..."
    
    # Create a simple AMI test
    (
        echo "Action: Login"
        echo "Username: $AMI_USER"
        echo "Secret: $AMI_PASS"
        echo ""
        echo "Action: Logoff"
        echo ""
        sleep 1
    ) | timeout 5 nc "$ASTERISK_HOST" "$AMI_PORT" > /tmp/ami_test.log 2>&1
    
    if grep -q "Authentication accepted" /tmp/ami_test.log 2>/dev/null; then
        echo "✅ AMI credentials are valid!"
    elif grep -q "Authentication failed" /tmp/ami_test.log 2>/dev/null; then
        echo "❌ AMI authentication failed - check AMI_USER and AMI_PASS"
    else
        echo "⚠️  Could not verify AMI credentials (this may be normal)"
    fi
    
    rm -f /tmp/ami_test.log
    
else
    echo "❌ Cannot connect to AMI port $AMI_PORT on $ASTERISK_HOST"
    echo "   Make sure Asterisk is running and AMI is enabled"
    echo ""
    echo "   To enable AMI, add this to /etc/asterisk/manager.conf:"
    echo "   [general]"
    echo "   enabled = yes"
    echo "   port = 5038"
    echo "   bindaddr = 0.0.0.0"
    echo ""
    echo "   [$AMI_USER]"
    echo "   secret = $AMI_PASS"
    echo "   read = all"
    echo "   write = all"
    echo ""
    echo "   Then run: sudo asterisk -rx 'manager reload'"
fi

echo ""

# Step 4: Build application
echo "🔨 Step 4: Build Application"
echo ""

if command -v go >/dev/null 2>&1; then
    echo "Building AMI webhook forwarder..."
    if go build -o ami-webhook-forwarder ami-webhook-forwarder.go; then
        echo "✅ Build successful!"
        
        echo ""
        echo "🎉 Setup Complete!"
        echo "=================="
        echo ""
        echo "Your configuration is saved in config.env"
        echo ""
        echo "To start the forwarder:"
        echo "  ./run-ami-forwarder.sh"
        echo ""
        echo "To test events:"
        echo "  ./test-events.sh"
        echo ""
        echo "Configuration summary:"
        echo "  Asterisk: $ASTERISK_HOST:$AMI_PORT"
        echo "  AMI User: $AMI_USER"
        echo "  Webhook:  $WEBHOOK_URL"
        
    else
        echo "❌ Build failed - check Go installation"
    fi
else
    echo "❌ Go not found - please install Go first"
    echo "   Visit: https://golang.org/dl/"
fi

echo ""
