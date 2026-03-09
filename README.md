# Shiori

A market news aggregator that scrapes Indonesian financial news sources.

## Features

- Real-time updates on news sources
- Focus on Indonesian market news (IHSG, stocks, bonds, commodities)
- Sources include major Indonesian financial news outlets

## Getting Started

### Prerequisites

- Go 1.25+
- Docker (optional, for containerized deployment)

### Running the Application

1. Start the server:
```bash
cd server
go run cmd/server/main.go
```

Or using Docker:
```bash
cd server
docker compose up --build
```

2. Stop the server:
```bash
# For Docker
docker compose down
```

### Development

```bash
# Run backend
cd server && go run cmd/server/main.go

# Run frontend
cd web && pnpm dev
```

## Disclaimer

This project is for educational purposes only. The content aggregated from third-party sources is not hosted or owned by this application. All credit for the news content belongs to the respective original publishers. This project does not intend to infringe on any copyrights or intellectual property rights.

## License

MIT
