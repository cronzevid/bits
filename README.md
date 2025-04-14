# Bits - Bitcoin Price Monitoring Bot

A Telegram bot that monitors Bitcoin (BTC) price changes on Bitstamp exchange and provides notifications for significant price movements.

## Features

- Real-time BTC price monitoring
- Price change notifications for different time periods:
  - Hourly (≥3% change)
  - Daily (≥5% change)
  - Weekly (≥10% change)
- Trading functionality through Bitstamp API
- User management and authentication
- Configurable notification thresholds

## Prerequisites

- Go 1.16 or higher
- Telegram Bot Token (from [@BotFather](https://t.me/BotFather))
- Bitstamp API credentials

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/bits.git
cd bits
```

2. Install dependencies:
```bash
go mod download
```

3. Configure the bot:
   - Copy `config.example.json` to `config.json`
   - Add your Telegram bot token
   - Add your Bitstamp API credentials

4. Build the bot:
```bash
go build -o bits
```

## Configuration

The bot uses the following configuration files:

- `config.json`: Main configuration file
  - `telegram_token`: Your Telegram bot token
  - `bitstamp_key`: Bitstamp API key
  - `bitstamp_secret`: Bitstamp API secret
  - `bitstamp_uid`: Bitstamp user ID

## Usage

Start the bot:
```bash
./bits
```

### Available Commands

- `/start` - Start receiving price notifications
- `/stop` - Stop receiving notifications
- `/register` - Register your Bitstamp API credentials
- `/balance` - Check your BTC and USD balance
- `/buy` - Buy BTC with available USD
- `/sell` - Sell all available BTC
- `/open` - Show open orders
- `/cancel` - Cancel all orders
- `/help` - Show this help message

## Project Structure

```
bits/
├── cmd/
│   └── bits/
│       └── main.go
├── internal/
│   ├── bot/         # Telegram bot implementation
│   ├── config/      # Configuration management
│   ├── exchange/    # Bitstamp API integration
│   ├── notifier/    # Price change notification logic
│   └── storage/     # Data persistence
├── pkg/            # Public packages
├── config.json     # Configuration file
└── README.md
```

## Error Handling

The bot implements comprehensive error handling:
- API call failures are logged and retried
- Invalid user input is handled gracefully
- Configuration errors are caught at startup
- Network issues are handled with retries

## Disclaimer

This bot is provided as-is. Use at your own risk. The developers are not responsible for any financial losses incurred while using this bot.
