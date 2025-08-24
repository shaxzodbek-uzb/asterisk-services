#!/bin/bash

# Test script to generate AMI events for webhook testing
# Run this to create test calls and see webhook events

echo "🧪 Testing AMI Webhook Forwarder"
echo "================================"

echo "📞 Generating test events using Asterisk CLI..."

# Test 1: Show channels (may trigger events)
echo "Test 1: Showing active channels..."
asterisk -rx "core show channels"

sleep 2

# Test 2: Originate a test call (creates real events)
echo "📞 Test 2: Creating test call..."
asterisk -rx "originate Local/test@default extension echo@default"

sleep 3

# Test 3: Show SIP peers (registration events)
echo "📞 Test 3: Checking SIP peers..."
asterisk -rx "sip show peers" 2>/dev/null || asterisk -rx "pjsip show endpoints" 2>/dev/null || echo "No SIP configured"

sleep 2

# Test 4: Show queues (if you have queue system)
echo "📞 Test 4: Checking queues..."
asterisk -rx "queue show" 2>/dev/null || echo "No queues configured"

sleep 2

# Test 5: Show parked calls
echo "📞 Test 5: Checking parked calls..."
asterisk -rx "parkedcalls show" 2>/dev/null || echo "No parked calls"

echo ""
echo "✅ Test events generated!"
echo ""
echo "📝 Check your AMI webhook forwarder logs for:"
echo "   - '📞 Received AMI event' messages"
echo "   - '✅ Successfully forwarded event' messages"
echo "   - 'AMI Heartbeat: Connection alive' messages"
echo ""
echo "💡 To generate MORE events:"
echo "   1. Make real calls between extensions"
echo "   2. Register/unregister SIP phones"
echo "   3. Transfer calls"
echo "   4. Put calls on hold"
echo "   5. Use features like *8 for voicemail"
echo ""
echo "🌐 Check your webhook URL for received events!"
