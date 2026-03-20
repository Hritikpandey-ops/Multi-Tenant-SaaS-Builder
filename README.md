# Multi-Tenant SaaS Platform

A production-ready multi-tenant SaaS platform built with **Go**, **PostgreSQL**, and **Redis**. Features microservices architecture with JWT authentication, Row-Level Security (RLS) for tenant isolation, and scalable design.

## 🏗️ Architecture

### Services
- **API Gateway** - Single entry point, JWT validation, request routing
- **Auth Service** - User authentication, tenant management, JWT issuance
- **Billing Service** - Subscription management, Stripe integration (planned)
- **Analytics Service** - Usage tracking, event analytics (planned)

### Tech Stack
| Component | Technology |
|-----------|-----------|
| Language | Go 1.21+ |
| Web Framework | Gin |
| Database | PostgreSQL 16 with RLS |
| Cache | Redis 7 |
| Message Queue | NATS (optional) |
| Metrics | Prometheus + Grafana |
| Deployment | Docker + Kubernetes |

## 📁 Project Structure

```
multi-tenant-saas/
├── cmd/                    # Service entrypoints
│   ├── gateway/           # API Gateway
│   ├── auth/              # Authentication Service
│   ├── billing/           # Billing Service (planned)
│   └── analytics/         # Analytics Service (planned)
├── internal/
│   ├── config/           # Configuration management
│   ├── database/         # Database connections (Postgres, Redis)
│   ├── handlers/         # HTTP handlers
│   ├── jwt/              # JWT management
│   ├── middleware/       # Gin middleware (auth, CORS, logging)
│   ├── repository/       # Data access layer
│   └── types/            # Type definitions
├── db/
│   └── init/            # Database initialization scripts
├── deployments/
│   ├── k8s/             # Kubernetes manifests
│   ├── prometheus/      # Prometheus config
│   └── grafana/         # Grafana dashboards
├── docker-compose.yml   # Local development
├── Dockerfile          # Multi-stage build
├── Makefile           # Build automation
├── go.mod
└── README.md
```

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 16+ (or use Docker)

### 1. Clone and Setup
```bash
# Copy environment variables
cp .env.example .env

# Edit .env with your configuration
```

### 2. Start Infrastructure
```bash
# Start PostgreSQL, Redis, NATS, Prometheus, Grafana
docker-compose up -d

# Verify services are running
docker-compose ps
```

### 3. Initialize Database
```bash
# The database will be auto-initialized on first start
# Or manually run:
psql -h localhost -U postgres -d multitenant -f db/init/01-init.sql
```

### 4. Run Services
```bash
# Install dependencies
go mod download

# Run auth service
make run/auth

# Run gateway (in another terminal)
make run/gateway

# Or run all services
make run
```

### 5. Test the API
```bash
# Register a new tenant and user
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "Hritik.Pandey@example.com",
    "password": "HritikPandey@24",
    "first_name": "Hritik",
    "last_name": "Pandey",
    "tenant_name": "Test Project",
    "tenant_slug": "test-project"
  }'

# Login
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "SecurePassword123!"
  }'
```

## 🔐 Authentication Flow

1. **Register**: Create tenant + first user (becomes OWNER)
2. **Login**: Receive JWT access token + refresh token
3. **Request**: Include `Authorization: Bearer <token>` header
4. **Gateway**: Validates JWT, extracts tenant context
5. **Service**: Receives request with tenant context
6. **Database**: RLS policies ensure tenant data isolation

## 🏢 Multi-Tenancy Model

### Pooled Model with RLS
- All tenants share the same database tables
- Each table has a `tenant_id` column
- **Row-Level Security (RLS)** policies enforce isolation
- Context is set per connection: `SET app.tenant_id = '<tenant_id>'`

### RLS Policy Example
```sql
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_policy ON users
  USING (tenant_id::text = current_setting('app.tenant_id'));
```

## 📊 Monitoring

Access monitoring dashboards:
- **Grafana**: http://localhost:3001 (admin/admin)
- **Prometheus**: http://localhost:9090

## 🔧 Development

### Available Commands
```bash
make build          # Build all services
make run            # Run all services
make test           # Run tests
make lint           # Run linter
make fmt            # Format code
make docker-up      # Start Docker services
make docker-down    # Stop Docker services
```

### Environment Variables
See [.env.example](.env.example) for all available configuration options.

## 🚢 Deployment

### Docker
```bash
# Build images
docker build -t multitenant-saas/gateway:latest .
docker build -t multitenant-saas/auth:latest .

# Or use docker-compose
docker-compose up -d
```

### Kubernetes
```bash
# Apply manifests
kubectl apply -f deployments/k8s/

# Check status
kubectl get pods -n multitenant-saas
kubectl get svc -n multitenant-saas
```

## 📚 API Documentation

### Authentication Endpoints
- `POST /api/v1/auth/register` - Register new tenant + user
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/refresh` - Refresh access token
- `GET /api/v1/auth/me` - Get current user (protected)

### User Management
- `GET /api/v1/users` - List users (protected)
- `GET /api/v1/users/:id` - Get user (protected)
- `PUT /api/v1/users/:id` - Update user (protected)
- `DELETE /api/v1/users/:id` - Delete user (protected)

### Tenant Management
- `GET /api/v1/tenant` - Get tenant info (protected)
- `PUT /api/v1/tenant` - Update tenant (protected, owner only)
- `GET /api/v1/tenant/users` - List tenant users (protected)
- `GET /api/v1/tenant/usage` - Get usage stats (protected)

## 🛡️ Security Features

- **JWT Authentication**: RS256 recommended for production
- **Password Hashing**: bcrypt with appropriate cost factor
- **Tenant Isolation**: PostgreSQL Row-Level Security
- **Rate Limiting**: Redis-based sliding window
- **CORS**: Configurable per environment
- **Request Logging**: Structured logging with tenant context

## 🗺️ Roadmap

- [ ] Billing Service with Stripe integration
- [ ] Analytics/Usage Service
- [ ] Email notifications (invite, password reset)
- [ ] OAuth2/OIDC support
- [ ] Multi-factor authentication
- [ ] API versioning
- [ ] WebSocket support
- [ ] Admin dashboard
- [ ] Automated backups
- [ ] Database sharding for scale

## 📄 License

MIT License - see LICENSE file for details

## 🤝 Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## 📞 Support

For issues and questions, please open a GitHub issue.
