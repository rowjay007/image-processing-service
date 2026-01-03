# Ports Specification

This document defines the core interfaces (Ports) of the Image Processing Service. These ports reside in the `internal/ports` package and decouple the business logic from infrastructure implementation details.

## 👥 Authentication & Identity

### `UserRepository`
Handles persistence of user identity data.
- `Create(ctx, user)`: Persists a new user record.
- `GetByID(ctx, id)`: Retrieves a user by their unique identifier.
- `GetByUsername(ctx, username)`: Retrieves a user by their username (for login).

### `AuthProvider`
Handles security token issuance and validation.
- `GenerateToken(userID, username)`: Creates a JWT for an authenticated user.
- `ValidateToken(token)`: Verifies a JWT and extracts claims.

## 🖼️ Image Management

### `ImageRepository`
Handles persistence of image metadata and variants.
- `Save(ctx, image)`: Persists image metadata.
- `GetByID(ctx, id)`: Retrieves image metadata by ID.
- `List(ctx, ownerID, offset, limit)`: Lists images owned by a user with pagination.
- `SaveVariant(ctx, imageID, variant)`: Persists metadata for a specific image transformation.
- `GetVariantBySpecHash(ctx, imageID, specHash)`: Retrieves a variant by its unique transformation signature.

### `ObjectStorage`
Abstracts binary data storage (e.g., Cloudinary, S3).
- `Put(ctx, key, reader, contentType, size)`: Uploads binary data and returns a URL/Key.
- `Get(ctx, key)`: Retrieves binary data as a readable stream.
- `SignedURL(ctx, key, expiry)`: Generates a temporary secure URL for direct access.
- `Delete(ctx, key)`: Removes binary data from storage.

## ⚙️ Processing & Caching

### `ImageProcessor`
Defines core image manipulation logic.
- `Transform(ctx, reader, spec)`: Applies a `TransformationSpec` to an image.
- `ExtractMetadata(ctx, reader)`: extracts width, height, and mime-type from raw bytes.

### `Cache`
Fast key-value storage for performance (e.g., Redis).
- `Get(ctx, key)`: Retrieves cached string data.
- `Set(ctx, key, value, ttl)`: Stores data with an expiration.
- `Incr(ctx, key)`: Increments an atomic counter (for rate limiting).
- `Expire(ctx, key, ttl)`: Sets expiration for an existing key.

## 📬 Communication

### `Queue`
Handles asynchronous job distribution (e.g., RabbitMQ).
- `Publish(ctx, job)`: Enqueues a transformation task for background processing.
- `Consume(ctx, handler)`: Listens for incoming jobs and processes them via the provided handler.

### `RateLimiter`
Handles traffic management.
- `Allow(ctx, key, limit, window)`: Checks if an action is permitted within a time window.
