# Fiber Backend API

A modern REST API for a blog/content management system built with Go, Fiber framework, and PostgreSQL. Features include user management, posts, comments, chat functionality, and more.

## Features

- **User Management**: Complete user system with authentication, profiles, and following
- **Content Management**: Posts with rich text, images, tags, and versioning
- **Social Features**: User follows, post likes, comments, and bookmarks
- **Real-time Chat**: Conversational AI with message history and token tracking
- **Analytics**: Post view tracking and statistics
- **Security**: JWT authentication, rate limiting, and input validation
- **Performance**: Database connection pooling, caching, and optimized queries
- **Deployment**: Docker support with multi-stage builds
- **Monitoring**: Comprehensive logging and metrics

## Quick Start

### Prerequisites
- Go 1.25+
- PostgreSQL 14+
- Docker (optional, for containerized deployment)

### Setup

1. **Clone and setup:**
   ```bash
   git clone <your-repo-url>
   cd fiberbackend
   cp .env.example .env
   ```

2. **Configure environment:**
   Edit `.env` file with your configuration:
   ```env
   # Server
   PORT=8080
   
   # Database
   DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable
   MAX_OPEN_CONNS=30
   MAX_IDLE_CONNS=2
   CONN_MAX_LIFETIME=30m
   
   # Authentication
   JWT_SECRET=your-secret-key

   # Rate Limiting
   RATE_LIMITER_MAX=0
   RATE_LIMITER_TTL=60

   # Debug
   DEBUG=false
   ```

3. **Run:**
   ```bash
   # Install dependencies
   go mod download
   
   # Run migrations (if using migration tool)
   # psql -d your_database_name -f migrations/*.sql
   
   # Start the server
   go run cmd/main.go
   ```

   Server starts at `http://localhost:8080`

### Development

```bash
# Build and run
make build
make dev

# Run tests
make test

# Code quality
make fmt
make vet
make lint
```

### Docker

```bash
# Build and run
docker build -t fiberbackend .
docker run -p 8080:8080 fiberbackend
```

## API Documentation

For detailed API endpoints and examples, see [api_doc.md](api_doc.md).

## Project Structure

```
├── cmd/                    # Application entry point
├── config/                 # Configuration management
├── internal/               # Private application code
│   ├── handler/            # HTTP handlers
│   ├── service/            # Business logic
│   ├── repository/         # Data access
│   ├── model/              # Database models
│   ├── middleware/         # HTTP middleware
│   └── routes/             # Route definitions
├── migrations/             # Database migrations
├── test/                   # Test files
└── pkg/                    # Shared utilities
```

This project follows a clean architecture pattern with separation of concerns between handlers, services, repositories, and models.

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Commit changes: `git commit -am 'Add feature'`
4. Push to branch: `git push origin feature/my-feature`
5. Submit a pull request

Please follow existing code style and include tests for new features.

## License

MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/your-repo/fiberbackend/issues)
- **Documentation**: [API Documentation](api_doc.md)

## Technologies

- **Go**: 1.25+
- **Fiber Framework**: v3.0.0
- **GORM**: v1.31.1
- **PostgreSQL**: 14+
- **JWT Authentication**: github.com/golang-jwt/jwt/v5
- **Docker**: Containerization support
