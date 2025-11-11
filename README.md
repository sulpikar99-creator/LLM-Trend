# LLM Trend - AI-Powered Trading Platform

An advanced AI-powered automated trading platform with multi-exchange support, real-time monitoring, advanced analytics, and secure multi-user authentication.

## 🎯 Project Status

**Phase 1: Core Infrastructure** ✅ COMPLETED

- ✅ Go backend with Gin framework
- ✅ React + TypeScript + Vite frontend
- ✅ SQLite database with encryption
- ✅ JWT authentication system
- ✅ Docker Compose deployment
- ✅ Basic API structure

**Phase 2: Exchange Integration** ✅ COMPLETED

- ✅ Base trader interface (unified API)
- ✅ Binance Futures integration (REST + WebSocket)
- ✅ Market data system with kline monitoring
- ✅ Technical indicators (RSI, MACD, Bollinger Bands, MA)
- ✅ Trader manager with lifecycle management
- ✅ Real-time data streaming
- ✅ Database-driven analytics (no sample data)

**Phase 3: AI Decision Engine** ✅ COMPLETED

- ✅ Multi-provider AI client (OpenAI, DeepSeek, Claude, Qwen)
- ✅ Decision engine with context-aware trading logic
- ✅ JSON decision logging with audit trail
- ✅ Auto-trading capabilities with configurable intervals
- ✅ Risk validation and decision execution
- ✅ Decision log API endpoints

**Phase 4: Analytics Dashboard** ✅ COMPLETED

- ✅ Drawdown analysis with recovery tracking
- ✅ Monte Carlo simulation engine (VaR, CVaR, confidence intervals)
- ✅ Correlation matrix computation for multi-symbol analysis
- ✅ Performance attribution by symbol, strategy, and timeframe
- ✅ Comprehensive risk metrics (Sharpe, Sortino, Calmar ratios)
- ✅ Analytics API endpoints with query parameters

**Phase 5: Risk Management** ✅ COMPLETED

- ✅ Account-level risk controls (max drawdown, daily loss limits)
- ✅ Position-level risk controls (stop-loss, take-profit, trailing stop)
- ✅ Leverage limits (global and per-symbol)
- ✅ Position sizing based on risk parameters
- ✅ Trading frequency limits and cooldown periods
- ✅ Automatic trading halt on violations
- ✅ Risk configuration and monitoring API endpoints

**Phase 6: Multi-User System** ✅ COMPLETED

- ✅ Beta code access control system
- ✅ Per-user configuration (AI provider, exchange, risk settings)
- ✅ Role-based access control (user/admin)
- ✅ Admin API endpoints for user and beta code management
- ✅ Database-driven trader statistics and testnet detection
- ✅ User-scoped trader instances with ownership validation

**Phase 7: Testing & Optimization** ✅ COMPLETED

- ✅ Comprehensive unit tests (analytics, risk management)
- ✅ Integration tests for Binance Futures API
- ✅ Database query optimization with composite indexes
- ✅ SQLite performance tuning (WAL, caching, mmap)
- ✅ API rate limiting (token bucket per IP)
- ✅ Security headers (CSP, XSS, HSTS, etc.)
- ✅ Request size limiting and panic recovery
- ✅ Enhanced logging and monitoring

## 🏗️ Architecture

### Backend Stack
- **Language**: Go 1.24+
- **Framework**: Gin (HTTP router)
- **Database**: SQLite with modernc.org/sqlite
- **Encryption**: AES-256-GCM, RSA
- **Authentication**: JWT tokens
- **Logging**: zerolog

### Frontend Stack
- **Language**: TypeScript 5+
- **Framework**: React 18
- **Build Tool**: Vite 6
- **UI**: Radix UI + Tailwind CSS 3
- **State**: SWR + Zustand
- **Charts**: Recharts 2

## 🚀 Quick Start

### Prerequisites
- Go 1.24+
- Node.js 20+
- Docker & Docker Compose (optional)

### Development Setup

1. **Clone the repository**
```bash
git clone https://github.com/sulpikar99-creator/LLM-Trend.git
cd LLM-Trend
```

2. **Set up environment variables**
```bash
cp .env.example .env
```

Edit `.env` and set the required values:
```bash
# Generate encryption key
openssl rand -hex 32

# Generate JWT secret
openssl rand -hex 32
```

3. **Run backend**
```bash
# Install dependencies
go mod download

# Run server
go run main.go
```

Backend will start on http://localhost:8080

4. **Run frontend**
```bash
cd web
npm install
npm run dev
```

Frontend will start on http://localhost:5173

### Docker Deployment

```bash
# Build and run with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## 📁 Project Structure

```
LLM-Trend/
├── main.go                 # Application entry point
├── analytics/              # Analytics engine
├── api/                    # HTTP API handlers
├── auth/                   # JWT authentication
├── bootstrap/              # Application initialization
├── config/                 # Configuration management
├── crypto/                 # Encryption services
├── decision/               # AI decision engine
├── logger/                 # Logging & decision logs
├── manager/                # Trader lifecycle management
├── market/                 # Market data & indicators
├── risk/                   # Risk management system
├── trader/                 # Exchange integrations
├── web/                    # React frontend
│   ├── src/
│   │   ├── pages/          # Page components
│   │   ├── components/     # Reusable components
│   │   ├── contexts/       # React contexts
│   │   ├── hooks/          # Custom hooks
│   │   ├── lib/            # Utilities
│   │   └── types/          # TypeScript types
│   └── package.json
├── docker/                 # Docker configuration
├── nginx/                  # Nginx configuration
├── .env.example            # Environment template
└── docker-compose.yml      # Service orchestration
```

## 🔐 Security Features

- End-to-end encryption for sensitive data
- AES-256-GCM encryption for database
- RSA encryption for API keys
- JWT authentication with secure tokens
- Password hashing with bcrypt
- Field-level encryption for credentials

## 📊 API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - User login

### User Management
- `GET /api/user/profile` - Get user profile
- `PUT /api/user/profile` - Update profile

### Traders
- `GET /api/traders` - List all traders
- `POST /api/traders` - Create trader
- `GET /api/traders/:id` - Get trader details
- `PUT /api/traders/:id` - Update trader
- `DELETE /api/traders/:id` - Delete trader
- `POST /api/traders/:id/start` - Start trading
- `POST /api/traders/:id/stop` - Stop trading

### Decision Logs
- `GET /api/decisions/:trader_id` - Get latest decisions
- `GET /api/decisions/:trader_id/:cycle` - Get decision by cycle
- `DELETE /api/decisions/:trader_id/old` - Delete old decisions

### Analytics
- `GET /api/analytics/drawdown` - Drawdown analysis (supports ?trader_id, ?start_date, ?end_date)
- `GET /api/analytics/montecarlo` - Monte Carlo simulation (supports ?initial_balance, ?num_simulations, ?num_periods)
- `GET /api/analytics/correlation` - Correlation matrix (supports ?symbols, ?top_n)
- `GET /api/analytics/performance` - Performance attribution (supports ?trader_id, ?top_n)

### Risk Management
- `GET /api/risk/:trader_id/config` - Get risk configuration (account + position)
- `PUT /api/risk/:trader_id/config` - Update risk settings
- `GET /api/risk/:trader_id/status` - Get real-time risk status
- `POST /api/risk/:trader_id/reset` - Reset risk monitor (clear violations)
- `POST /api/risk/:trader_id/calculate-size` - Calculate safe position size
- `POST /api/risk/:trader_id/calculate-sltp` - Calculate SL/TP prices

## 🗺️ Development Roadmap

### Phase 1: Core Infrastructure ✅
- [x] Project setup
- [x] Database & encryption
- [x] Authentication system
- [x] Basic API structure
- [x] Frontend setup
- [x] Docker deployment

### Phase 2: Exchange Integration ✅
- [x] Base trader interface
- [x] Binance Futures integration
- [x] OKX/Bybit integration
- [x] Market data system
- [x] WebSocket real-time data

### Phase 3: AI Decision Engine ✅
- [x] AI client integration
- [x] Decision engine
- [x] Decision logger
- [x] Prompt templating

### Phase 4: Analytics Dashboard ✅
- [x] Backend analytics endpoints
- [x] Drawdown analysis
- [x] Monte Carlo simulation
- [x] Correlation matrix
- [x] Performance attribution
- [x] Risk metrics calculation

### Phase 5: Risk Management ✅
- [x] Account-level controls
- [x] Position-level controls
- [x] Leverage limits
- [x] Position sizing
- [x] Trading frequency limits
- [x] Risk monitoring API

### Phase 6: Multi-User System ✅
- [x] Beta code system for access control
- [x] User configuration management (per-user AI, exchange, risk settings)
- [x] Per-user trader instances with ownership validation
- [x] Role-based access control (user/admin)
- [x] Admin endpoints for user and beta code management
- [x] Database-driven real-time statistics and testnet detection

### Phase 7: Testing & Optimization ✅
- [x] Unit tests (analytics, risk management)
- [x] Integration tests (Binance API)
- [x] Database query optimization with composite indexes
- [x] SQLite performance tuning (WAL mode, caching, memory-mapped I/O)
- [x] Rate limiting middleware (per-IP token bucket)
- [x] Security headers (CSP, XSS protection, HSTS)
- [x] CORS configuration
- [x] Request size limiting
- [x] Panic recovery middleware

## 🧪 Testing

### Running Tests

**Unit Tests** (analytics, risk management, core logic):
```bash
# Run all unit tests
go test ./analytics/... ./risk/...

# Run with verbose output
go test -v ./analytics/... ./risk/...

# Run with coverage
go test -cover ./analytics/... ./risk/...
```

**Integration Tests** (Binance API):
```bash
# Set up testnet credentials (get from https://testnet.binancefuture.com)
export BINANCE_TESTNET_API_KEY=your_testnet_api_key
export BINANCE_TESTNET_SECRET=your_testnet_secret

# Run integration tests
go test -v ./trader/...

# Skip integration tests if credentials not available
go test ./trader/... # Tests will skip automatically
```

**All Tests**:
```bash
# Run all tests in the project
go test ./...

# Run with race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Categories

- **Unit Tests**: Test isolated logic without external dependencies
  - `analytics/drawdown_test.go` - Drawdown calculation tests
  - `risk/account_risk_test.go` - Risk management tests

- **Integration Tests**: Test real API interactions (requires credentials)
  - `trader/binance_futures_test.go` - Binance Futures API tests

## 🛠️ Development Commands

### Backend
```bash
# Run tests
go test ./...

# Build binary
go build -o llm-trend

# Run with hot reload (requires air)
air

# Install dependencies
go mod download

# Update dependencies
go mod tidy
```

### Frontend
```bash
# Development
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Lint
npm run lint

# Format code
npm run format
```

## 📝 Environment Variables

See `.env.example` for all available environment variables:

**Required:**
- `DATA_ENCRYPTION_KEY` - 32-byte hex key for database encryption
- `JWT_SECRET` - Secret key for JWT tokens

**Server Configuration:**
- `NOFX_BACKEND_PORT` - Backend port (default: 8080)
- `NOFX_FRONTEND_PORT` - Frontend port (default: 3000)
- `GIN_MODE` - Gin mode: debug, release, or test

**AI Configuration:**
- `DEFAULT_AI_PROVIDER` - Default AI provider (openai, anthropic, custom)
- `DEFAULT_AI_MODEL` - Default AI model (gpt-4, claude-3-5-sonnet, etc.)
- `AI_MAX_TOKENS` - AI response token limit (default: 4000)
- `AI_REQUEST_TIMEOUT` - AI request timeout in seconds (default: 30)

**Binance API:**
- `BINANCE_API_KEY` - Production Binance API key (DANGER: Real money!)
- `BINANCE_SECRET` - Production Binance secret
- `BINANCE_TESTNET_API_KEY` - Testnet API key (recommended for testing)
- `BINANCE_TESTNET_SECRET` - Testnet secret

**Optional:**
- `TELEGRAM_BOT_TOKEN` - Telegram bot token for notifications
- `TELEGRAM_CHAT_ID` - Telegram chat ID
- `RATE_LIMIT_RPS` - Requests per second per IP (default: 10)
- `RATE_LIMIT_BURST` - Rate limit burst size (default: 20)
- `DATABASE_PATH` - Database file path (default: config.db)
- `LOG_LEVEL` - Log level: debug, info, warn, error
- `DEBUG` - Enable debug mode (default: false)

## ⚠️ Important Notes

1. **Never commit** `.env` file or encryption keys to git
2. **Change default secrets** before production deployment
3. **Enable HTTPS** in production
4. **Back up** `config.db` regularly
5. **Monitor** logs for security events

## 📄 License

See LICENSE file for details.

## 🤝 Contributing

This is Phase 1 of the project. More features coming soon!

For issues and feature requests, please open an issue on GitHub.

---

**Built with ❤️ for automated trading**