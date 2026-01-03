# Image Processing Service 🚀

This project is based on the [Image Processing Service](https://roadmap.sh/projects/image-processing-service) project from roadmap.sh. It is a high-performance, scalable image processing service built with Go, leveraging Hexagonal Architecture and DDD principles. It provides asynchronous image transformations, secure storage, and efficient caching.

## ✨ Core Features

- **Asynchronous Processing**: Uses RabbitMQ (CloudAMQP) to handle heavy image transformations in the background.
- **Secure Authentication**: JWT-based authentication with Bcrypt password hashing.
- **Scalable Storage**: Cloud-native storage integration with Cloudinary.
- **Lightning Fast Caching**: Layered caching with Upstash Redis for metadata and image variants.
- **Robust Persistence**: Supabase (Postgres) for reliable metadata storage.
- **Flexible Transformations**: (WIP) resizing, cropping, and format conversion.

## 🏗️ Architecture

The project follows **Hexagonal Architecture (Ports and Adapters)** combined with **Domain-Driven Design (DDD)**:

- `internal/domain`: Pure business logic and entities (Image, User, Variant).
- `internal/ports`: Interface definitions for infrastructure (Storage, DB, Cache, Queue).
- `internal/adapters`: Concrete implementations (Cloudinary, Postgres, Redis, RabbitMQ).
- `internal/application`: Use Cases coordinating domain logic and ports.
- `cmd/api`: Entry point for the RESTful API server.
- `cmd/worker`: Entry point for the background job consumer.

## 🚀 Quick Start

### Prerequisites
- Go 1.24+
- [Task](https://taskfile.dev/) (optional, but recommended)
- Cloud accounts for Supabase, Cloudinary, Upstash, and CloudAMQP.

### Setup
1. Clone the repository.
2. `cp .env.example .env` and fill in your credentials.
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Run migrations:
   ```bash
   task db:migrate
   ```

### Running the System
- **API Server**: `task run:api` (Runs on port 8080)
- **Worker**: `task run:worker`

## 📚 Documentation

Detailed documentation is available in the `docs/` folder:

- [**Development Guide**](docs/development.md) - Local setup, testing, and task runner.
- [**API Reference**](docs/api.md) - Endpoint documentation and authentication.
- [**Deployment Guide**](docs/deployment.md) - Production considerations and infra setup.

## 🧪 Testing

We use both unit and integration tests to ensure system stability.

- Run all tests: `task test`
- Run integration tests: `go test -v test/integration/live_test.go`
- Verify async flow: `sh verify_worker.sh`

## 📄 License
MIT
