package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"image-processing-service/internal/domain/image"
	"image-processing-service/internal/domain/user"
)

type PostgresImageRepository struct {
	db *pgxpool.Pool
}

func NewPostgresImageRepository(db *pgxpool.Pool) *PostgresImageRepository {
	return &PostgresImageRepository{
		db: db,
	}
}

func (r *PostgresImageRepository) Save(ctx context.Context, img *image.Image) error {
	query := `
		INSERT INTO images (id, owner_id, filename, original_key, size, mime_type, width, height, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		img.ID,
		img.OwnerID,
		img.Filename,
		img.OriginalKey,
		img.Size,
		img.MimeType,
		img.Width,
		img.Height,
		img.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}
	return nil
}

func (r *PostgresImageRepository) SaveVariant(ctx context.Context, imageID image.ImageID, variant *image.Variant) error {
	query := `
		INSERT INTO variants (id, image_id, variant_key, spec_hash, size, mime_type, width, height, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (image_id, spec_hash) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query,
		variant.ID,
		imageID,
		variant.VariantKey,
		variant.SpecHash,
		variant.Size,
		variant.MimeType,
		variant.Width,
		variant.Height,
		variant.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save variant: %w", err)
	}
	return nil
}

func (r *PostgresImageRepository) GetByID(ctx context.Context, id image.ImageID) (*image.Image, error) {
	imgQuery := `
		SELECT id, owner_id, filename, original_key, size, mime_type, width, height, created_at
		FROM images
		WHERE id = $1
	`
	row := r.db.QueryRow(ctx, imgQuery, id)

	var img image.Image
	var idStr, ownerIDStr string
	err := row.Scan(
		&idStr,
		&ownerIDStr,
		&img.Filename,
		&img.OriginalKey,
		&img.Size,
		&img.MimeType,
		&img.Width,
		&img.Height,
		&img.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get image: %w", err)
	}
	img.ID = image.ImageID(idStr)
	img.OwnerID = user.UserID(ownerIDStr)

	varQuery := `
		SELECT id, variant_key, spec_hash, size, mime_type, width, height, created_at
		FROM variants
		WHERE image_id = $1
	`
	rows, err := r.db.Query(ctx, varQuery, idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get variants: %w", err)
	}
	defer rows.Close()

	img.Variants = make([]image.Variant, 0)
	for rows.Next() {
		var v image.Variant
		var vId uuid.UUID
		if err := rows.Scan(
			&vId,
			&v.VariantKey,
			&v.SpecHash,
			&v.Size,
			&v.MimeType,
			&v.Width,
			&v.Height,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		v.ID = vId
		img.Variants = append(img.Variants, v)
	}

	return &img, nil
}

func (r *PostgresImageRepository) List(ctx context.Context, ownerID user.UserID, offset, limit int) ([]*image.Image, int, error) {
	// Count total
	countQuery := `SELECT COUNT(*) FROM images WHERE owner_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, ownerID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fetch items
	listQuery := `
		SELECT id, owner_id, filename, original_key, size, mime_type, width, height, created_at
		FROM images
		WHERE owner_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, listQuery, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	images := make([]*image.Image, 0)
	for rows.Next() {
		var img image.Image
		var idStr, ownerIDStr string
		if err := rows.Scan(
			&idStr,
			&ownerIDStr,
			&img.Filename,
			&img.OriginalKey,
			&img.Size,
			&img.MimeType,
			&img.Width,
			&img.Height,
			&img.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		img.ID = image.ImageID(idStr)
		img.OwnerID = user.UserID(ownerIDStr)
		images = append(images, &img)
	}

	return images, total, nil
}

func (r *PostgresImageRepository) GetVariantBySpecHash(ctx context.Context, imageID image.ImageID, specHash string) (*image.Variant, error) {
	query := `
		SELECT id, variant_key, spec_hash, size, mime_type, width, height, created_at
		FROM variants
		WHERE image_id = $1 AND spec_hash = $2
	`
	row := r.db.QueryRow(ctx, query, imageID, specHash)

	var v image.Variant
	var vId uuid.UUID
	err := row.Scan(
		&vId,
		&v.VariantKey,
		&v.SpecHash,
		&v.Size,
		&v.MimeType,
		&v.Width,
		&v.Height,
		&v.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get variant: %w", err)
	}
	v.ID = vId
	return &v, nil
}
