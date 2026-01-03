# API Reference

Detailed specifications for all ports and interfaces can be found in the [Ports Specification](architecture/ports.md).
Detailed database design can be found in the [Database Design](architecture/database.md).

All API endpoints are prefixed with `/api/v1` (except `/health`).

## Table of Contents
- [Authentication](#authentication)
  - [Register](#register)
  - [Login](#login)
  - [Get Profile](#get-profile)
- [Image Management](#image-management)
  - [Upload Image](#upload-image)
  - [Get Image Details](#get-image-details)
  - [List My Images](#list-my-images)
  - [Async Transform](#async-transform)
- [Miscellaneous](#miscellaneous)
  - [Health Check](#health-check)

---

## Authentication

### Register
`POST /auth/register`

Create a new user account.

**Request Body:**
```json
{
    "username": "johndoe",
    "password": "securepassword123"
}
```

### Login
`POST /auth/login`

Authenticate and receive a JWT.

**Request Body:**
```json
{
    "username": "johndoe",
    "password": "securepassword123"
}
```

**Response:**
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5..."
}
```

### Get Profile
`GET /me`

Get details of the currently authenticated user.
*Requires Authorization header: `Bearer <token>`*

---

## Image Management

### Upload Image
`POST /images`

Upload a raw image file. Metadata will be extracted automatically.
*Requires Authorization header: `Bearer <token>`*

**Content-Type:** `multipart/form-data`
**Parameters:**
- `file`: The actual image file.

**Response:**
```json
{
    "id": "uuid-v4",
    "filename": "original.jpg",
    "mime_type": "image/jpeg",
    "size": 10245,
    "width": 1920,
    "height": 1080,
    "created_at": "2026-01-01T12:00:00Z"
}
```

### Get Image Details
`GET /images/:id`

Fetch metadata and variants for a specific image.
*Requires Authorization header: `Bearer <token>`*

### List My Images
`GET /images?offset=0&limit=20`

List all images owned by the user with pagination.
*Requires Authorization header: `Bearer <token>`*

### Async Transform
`POST /images/:id/transform`

Schedule an asynchronous image transformation. Returns a `variant_id` that can be tracked.
*Requires Authorization header: `Bearer <token>`*

**Request Body:**
```json
{
    "resize": {
        "width": 100,
        "height": 100
    },
    "crop": {
        "x": 0,
        "y": 0,
        "width": 50,
        "height": 50
    },
    "format": "webp",
    "quality": 80
}
```

**Response:**
```json
{
    "message": "Transformation accepted",
    "variant_id": "uuid-v4"
}
```

---

## Miscellaneous

### Health Check
`GET /health`

Verify the service status.

**Response:**
```json
{
    "status": "ok",
    "time": "2026-01-01T12:00:00Z"
}
```
