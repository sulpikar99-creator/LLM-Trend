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
├── api/                    # HTTP API handlers
├── auth/                   # JWT authentication
├── bootstrap/              # Application initialization
├── config/                 # Configuration management
├── crypto/                 # Encryption services
├── logger/                 # Logging system
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

### Traders (Coming in Phase 2)
- `GET /api/traders` - List all traders
- `POST /api/traders` - Create trader
- `GET /api/traders/:id` - Get trader details
- `PUT /api/traders/:id` - Update trader
- `DELETE /api/traders/:id` - Delete trader
- `POST /api/traders/:id/start` - Start trading
- `POST /api/traders/:id/stop` - Stop trading

### Analytics (Coming in Phase 4)
- `GET /api/analytics/drawdown` - Drawdown analysis
- `GET /api/analytics/montecarlo` - Monte Carlo simulation
- `GET /api/analytics/correlation` - Correlation matrix
- `GET /api/analytics/performance` - Performance metrics

## 🗺️ Development Roadmap

### Phase 1: Core Infrastructure ✅
- [x] Project setup
- [x] Database & encryption
- [x] Authentication system
- [x] Basic API structure
- [x] Frontend setup
- [x] Docker deployment

### Phase 2: Exchange Integration (Next)
- [ ] Base trader interface
- [ ] Binance Futures integration
- [ ] OKX/Bybit integration
- [ ] Market data system
- [ ] WebSocket real-time data

### Phase 3: AI Decision Engine
- [ ] AI client integration
- [ ] Decision engine
- [ ] Decision logger
- [ ] Prompt templating

### Phase 4: Analytics Dashboard
- [ ] Backend analytics endpoints
- [ ] Frontend charts
- [ ] Risk monitoring

### Phase 5: Risk Management
- [ ] Account-level controls
- [ ] Position-level controls

### Phase 6: Multi-User System
- [ ] User management
- [ ] Trader management
- [ ] Role-based access

### Phase 7: Testing & Optimization
- [ ] Unit tests
- [ ] Integration tests
- [ ] Performance optimization
- [ ] Security audit

## 🛠️ Development Commands

### Backend
```bash
# Run tests
go test ./...

# Build binary
go build -o llm-trend

# Run with hot reload (requires air)
air
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

- `DATA_ENCRYPTION_KEY` - 32-byte hex key for database encryption (required)
- `JWT_SECRET` - Secret key for JWT tokens (required)
- `NOFX_BACKEND_PORT` - Backend port (default: 8080)
- `NOFX_FRONTEND_PORT` - Frontend port (default: 3000)
- `AI_MAX_TOKENS` - AI response token limit (default: 4000)

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