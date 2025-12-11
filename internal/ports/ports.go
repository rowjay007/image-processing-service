package ports

import (
	"context"
	"io"
	"time"

	"image-processing-service/internal/domain/image"
	"image-processing-service/internal/domain/user"
)

// UserRepository defines persistence operations for users.
type UserRepository interface {
	Create(ctx context.Context, u *user.User) error
	GetByID(ctx context.Context, id user.UserID) (*user.User, error)
	GetByUsername(ctx context.Context, username string) (*user.User, error)
}

// ImageRepository defines persistence operations for images and variants.
type ImageRepository interface {
	Save(ctx context.Context, img *image.Image) error
	GetByID(ctx context.Context, id image.ImageID) (*image.Image, error)
	List(ctx context.Context, ownerID user.UserID, offset, limit int) ([]*image.Image, int, error)
	SaveVariant(ctx context.Context, imageID image.ImageID, variant *image.Variant) error
	GetVariantBySpecHash(ctx context.Context, imageID image.ImageID, specHash string) (*image.Variant, error)
}

// ObjectStorage defines operations for storing and retrieving binary objects.
type ObjectStorage interface {
	Put(ctx context.Context, key string, reader io.Reader, contentType string, size int64) (string, error)
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	SignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}

// ProcessedImage represents the result of an image transformation.
type ProcessedImage struct {
	Data     []byte
	MimeType string
	Width    int
	Height   int
	Size     int64
}

// ImageMetadata represents basic metadata extracted from an image.
type ImageMetadata struct {
	Width    int
	Height   int
	MimeType string
	Size     int64
}

// ImageProcessor defines operations for transforming images.
type ImageProcessor interface {
	Transform(ctx context.Context, srcReader io.Reader, spec *image.TransformationSpec) (*ProcessedImage, error)
	ExtractMetadata(ctx context.Context, reader io.Reader) (*ImageMetadata, error)
}

// Cache defines operations for temporary key-value storage.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// TransformJob represents an asynchronous image transformation task.
type TransformJob struct {
	JobID    string                    `json:"job_id"`
	ImageID  image.ImageID             `json:"image_id"`
	OwnerID  user.UserID               `json:"owner_id"`
	Spec     *image.TransformationSpec `json:"spec"`
	SpecHash string                    `json:"spec_hash"`
}

// Queue defines operations for asynchronous job processing.
type Queue interface {
	Publish(ctx context.Context, job *TransformJob) error
	Consume(ctx context.Context, handler func(*TransformJob) error) error
}

// AuthProvider defines operations for token management.
type AuthProvider interface {
	GenerateToken(userID user.UserID, username string) (string, error)
	ValidateToken(token string) (*Claims, error)
}

// Claims represents the data embedded in a JWT.
type Claims struct {
	UserID   string
	Username string
}
