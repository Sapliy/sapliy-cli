# Sapliy CLI

[![Go Report Card](https://goreportcard.com/badge/github.com/sapliy/sapliy-cli)](https://goreportcard.com/report/github.com/sapliy/sapliy-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Developer CLI for the Sapliy Fintech Ecosystem. Login, listen to webhooks, trigger events, and debug flows — all from your terminal.

## Features

- **Authentication** — Login with your Sapliy account
- **Webhook Listening** — Stream events without tunnels (ngrok-free)
- **Event Triggering** — Fire test events from CLI
- **Zone Switching** — Switch between test/live modes
- **Flow Debugging** — View flow execution in real-time

## Installation

### macOS / Linux

```bash
# Using Homebrew
brew install sapliy/tap/sapliy

# Or download directly
curl -L https://github.com/sapliy/sapliy-cli/releases/download/latest/sapliy-$(uname -s)-$(uname -m) -o sapliy
chmod +x sapliy
sudo mv sapliy /usr/local/bin/

# Local Installation (from source)
# If you have the repository cloned locally:
make build
sudo make install
```

### From Source

```bash
go install github.com/sapliy/sapliy-cli/cmd/sapliy@latest
```

## Quick Start

```bash
# 1. Login to your account
sapliy login

# 2. Select your zone
sapliy zones list
sapliy zones use zone_abc123

# 3. Listen for webhooks locally
sapliy listen --forward-to http://localhost:4242/webhook

# 4. Trigger a test event (in another terminal)
sapliy trigger payment.succeeded --data '{"amount": 2000}'
```

## Commands

### Authentication

```bash
# Login (opens browser for OAuth)
sapliy login

# Check current session
sapliy whoami

# Logout
sapliy logout
```

### Zones

```bash
# List all zones
sapliy zones list

# Switch to a zone
sapliy zones use <zone_id>

# Show current zone
sapliy zones current

# Switch between test/live mode
sapliy mode test
sapliy mode live
```

### Webhook Listening

```bash
# Listen and print events to console
sapliy listen

# Forward to a local server
sapliy listen --forward-to http://localhost:4242/webhook

# Forward specific event types only
sapliy listen --events payment.succeeded,payment.failed --forward-to http://localhost:4242

# Show event payload
sapliy listen --print-json
```

### Triggering Events

```bash
# Trigger a test event
sapliy trigger payment.succeeded

# Trigger with custom data
sapliy trigger checkout.completed --data '{"cart_id": "cart_123", "total": 5000}'

# Trigger from a JSON file
sapliy trigger payment.created --file ./test-event.json
```

### Flows

```bash
# List flows in current zone
sapliy flows list

# Get flow details
sapliy flows get <flow_id>

# View recent flow executions
sapliy flows logs <flow_id>

# Enable/disable a flow
sapliy flows enable <flow_id>
sapliy flows disable <flow_id>
```

### Logs

```bash
# Stream recent events
sapliy logs

# Filter by event type
sapliy logs --type payment.succeeded

# Show last N events
sapliy logs --limit 50
```

## Configuration

The CLI stores configuration in `~/.sapliy/`:

```
~/.sapliy/
├── config.json    # Settings and preferences
├── credentials    # OAuth tokens (encrypted)
└── zones.json     # Zone cache
```

### Custom API Endpoint

For self-hosted deployments:

```bash
# Set custom endpoint
sapliy config set api_url https://api.yourdomain.com

# Or via environment variable
export SAPLIY_API_URL=https://api.yourdomain.com
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `SAPLIY_API_URL` | API endpoint (default: api.sapliy.io) |
| `SAPLIY_API_KEY` | API key for non-interactive use |
| `SAPLIY_ZONE` | Default zone ID |

## Local Development Workflow

A typical development session:

```bash
# Terminal 1: Start your local server
npm run dev  # Your app on :4242

# Terminal 2: Listen for webhooks
sapliy login
sapliy zones use zone_test_abc
sapliy listen --forward-to http://localhost:4242/webhook

# Terminal 3: Trigger test events
sapliy trigger payment.succeeded --data '{"amount": 1000}'
```

## Part of Sapliy Fintech Ecosystem

- [fintech-ecosystem](https://github.com/Sapliy/fintech-ecosystem) — Core backend
- [fintech-sdk-node](https://github.com/Sapliy/fintech-sdk-node) — Node.js SDK
- [fintech-automation](https://github.com/Sapliy/fintech-automation) — Flow Builder

## License

MIT © [Sapliy](https://github.com/sapliy)
