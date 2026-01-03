# Development Guide

This guide covers everything you need to know to develop, test, and contribute to the Image Processing Service.

## 🛠️ Task Runner

We use [Task](https://taskfile.dev/) to manage common development tasks. 

| Command | Description |
|---------|-------------|
| `task run:api` | Build and run the API server. |
| `task run:worker` | Build and run the Background Worker. |
| `task db:migrate` | Apply all pending SQL migrations. |
| `task test` | Run all units tests. |
| `task build` | Build both API and Worker binaries. |
| `task clean` | Remove build artifacts. |

## 🧪 Testing Strategy

### 1. Integration Tests
The most comprehensive way to verify the system is via the "Live Integration" test, which performs a full user lifecycle (Register -> Login -> Upload -> Retrieval).

**Requirement**: Both the API and Worker (optional for retrieval) must be running.

```bash
go test -v test/integration/live_test.go
```

### 2. Manual Verification
You can use the provided [Auth/Upload verification script](../verify_worker.sh) to trigger the full async pipeline:

```bash
sh verify_worker.sh
```

### 3. API Client
We include an `api_requests.http` file for use with the IntelliJ/VSCode HTTP Client extension (REST Client). This is the best way to interact with the API during development.

## 📂 Project Structure Explained

- **`cmd/`**: entry points (main functions).
- **`internal/domain/`**: Pure entities. No dependencies allowed here.
- **`internal/ports/`**: Interfaces defining what the application needs (DB, Storage, etc).
- **`internal/adapters/`**: How we talk to the outside world (Cloudinary adapter, Postgres adapter).
- **`internal/application/`**: Orchestration logic (Use Cases).
- **`internal/container/`**: Dependency Injection wiring.

## 🚧 Known Issues & Limitations

- **Processor**: Current implementation uses `StdLibImageProcessor` which only extracts metadata. Real transformations require `libvips` and the `bimg` adapter (WIP).
- **Redis Cache**: TTL is currently hardcoded to 1 hour in Use Cases.

## 📮 Contribution Guidelines

1. Fork the repo.
2. Create a feature branch.
3. Ensure `go fmt`, `go vet`, and `golangci-lint` pass.
4. Add tests for new functionality.
5. Submit a PR.
