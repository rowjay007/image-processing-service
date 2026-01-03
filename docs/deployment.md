# Deployment Guide

This guide outlines the steps and considerations for deploying the Image Processing Service to production.

## 🧱 Production Infrastructure Needs

The service depends on several external managed services:

1. **Database**: Postgres (Supabase recommended).
2. **Cache**: Redis (Upstash recommended).
3. **Queue**: RabbitMQ (CloudAMQP recommended).
4. **Storage**: Cloudinary.
5. **Runtime**: A containerized environment (Docker/Kubernetes/Render/Railway).

## 🔐 Secrets Management

Do **not** commit `.env` to version control. Use your platform's secret manager (e.g., GitHub Secrets, Kubernetes Secrets, Railway Environment Variables).

### Required Secrets Checklist:
- `SUPABASE_DB_URL`: Secure Postgres DSN.
- `UPSTASH_REDIS_PASSWORD`: Redis credentials.
- `CLOUDINARY_API_SECRET`: Storage credentials.
- `CLOUDAMQP_URL`: Queue connection string.
- `JWT_SECRET`: A long, random string (min 32 chars).

## 🚀 Deployment Steps

### 1. Build Binaries
The service produces two separate binaries. It is recommended to run them in separate containers/processes.

```bash
# Build API
go build -o build/api cmd/api/main.go

# Build Worker
go build -o build/worker cmd/worker/main.go
```

### 2. Database Migrations
Run migrations before starting the API server to ensure the schema is up to date:

```bash
# In production, you might want to run this as a standalone job/init-container
go run cmd/migrate/main.go
```

### 3. Execution

**API Server**:
- Expose port `8080`.
- Scale vertically or horizontally based on request volume.

**Worker**:
- Does not expose ports.
- Scale horizontally based on queue depth (Prefetch count is configurable).

## 📈 Scalability Considerations

- **Horizontal Scaling**: The API is stateless and can be scaled indefinitely.
- **Worker Scaling**: Increase the number of workers to handle high volumes of transformations. Ensure `DB_MAX_CONNS` and `QUEUE_PREFETCH_COUNT` are tuned accordingly.
- **Caching**: The Redis cache is critical for performance. Monitor Upstash memory usage.

## 🛡️ Security

- Enable **TLS** for all external connections (Redis, RabbitMQ, DB).
- Use a strong **JWT Secret**.
- Set `GIN_MODE=release` to disable debug logging.
- Limit max upload size (`MAX_UPLOAD_SIZE`) to prevent DOS.
