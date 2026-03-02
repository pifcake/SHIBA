package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"SHIBA/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ModelRepository struct {
	pool *pgxpool.Pool
}

func NewModelRepository(pool *pgxpool.Pool) *ModelRepository {
	return &ModelRepository{pool: pool}
}

func (r *ModelRepository) Create(ctx context.Context, p *domain.ModelProfile) error {
	q := `INSERT INTO model_profiles (id, user_id, first_name, last_name, birth_date, city, country,
		  willing_to_relocate, height_cm, weight_kg, chest_cm, waist_cm, hips_cm, shoe_size,
		  clothing_size, gender, hair_color, hair_length, hair_structure, eye_color,
		  clothing_size_top, clothing_size_bot, phone, bio, shoot_restrictions)
		  VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25)`
	_, err := r.pool.Exec(ctx, q,
		p.ID, p.UserID, p.FirstName, p.LastName, p.BirthDate, p.City, p.Country,
		p.WillingToRelocate, p.HeightCm, p.WeightKg, p.ChestCm, p.WaistCm, p.HipsCm, p.ShoeSize,
		p.ClothingSize, p.Gender, p.HairColor, p.HairLength, p.HairStructure, p.EyeColor,
		p.ClothingSizeTop, p.ClothingSizeBot, p.Phone, p.Bio, p.ShootRestrictions)
	if err != nil {
		return fmt.Errorf("create model profile: %w", err)
	}
	return nil
}

func (r *ModelRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.ModelProfile, error) {
	q := `SELECT id, user_id, first_name, last_name, birth_date, city, country,
		  willing_to_relocate, height_cm, weight_kg, chest_cm, waist_cm, hips_cm, shoe_size,
		  clothing_size, gender, hair_color, hair_length, hair_structure, eye_color,
		  clothing_size_top, clothing_size_bot, phone, bio, shoot_restrictions, created_at, updated_at
		  FROM model_profiles WHERE user_id = $1`
	row := r.pool.QueryRow(ctx, q, userID)
	return scanModelProfile(row)
}

func (r *ModelRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ModelProfile, error) {
	q := `SELECT id, user_id, first_name, last_name, birth_date, city, country,
		  willing_to_relocate, height_cm, weight_kg, chest_cm, waist_cm, hips_cm, shoe_size,
		  clothing_size, gender, hair_color, hair_length, hair_structure, eye_color,
		  clothing_size_top, clothing_size_bot, phone, bio, shoot_restrictions, created_at, updated_at
		  FROM model_profiles WHERE id = $1`
	row := r.pool.QueryRow(ctx, q, id)
	return scanModelProfile(row)
}

func (r *ModelRepository) Update(ctx context.Context, p *domain.ModelProfile) error {
	q := `UPDATE model_profiles SET
		  first_name=$1, last_name=$2, birth_date=$3, city=$4, country=$5,
		  willing_to_relocate=$6, height_cm=$7, weight_kg=$8, chest_cm=$9, waist_cm=$10,
		  hips_cm=$11, shoe_size=$12, clothing_size=$13,
		  gender=$14, hair_color=$15, hair_length=$16, hair_structure=$17, eye_color=$18,
		  clothing_size_top=$19, clothing_size_bot=$20,
		  phone=$21, bio=$22, shoot_restrictions=$23, updated_at=NOW()
		  WHERE id=$24`
	_, err := r.pool.Exec(ctx, q,
		p.FirstName, p.LastName, p.BirthDate, p.City, p.Country,
		p.WillingToRelocate, p.HeightCm, p.WeightKg, p.ChestCm, p.WaistCm,
		p.HipsCm, p.ShoeSize, p.ClothingSize,
		p.Gender, p.HairColor, p.HairLength, p.HairStructure, p.EyeColor,
		p.ClothingSizeTop, p.ClothingSizeBot,
		p.Phone, p.Bio, p.ShootRestrictions, p.ID)
	return err
}

// Categories

func (r *ModelRepository) AddCategory(ctx context.Context, profileID uuid.UUID, categoryID int16) error {
	q := `INSERT INTO model_categories (model_profile_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, q, profileID, categoryID)
	return err
}

func (r *ModelRepository) RemoveCategory(ctx context.Context, profileID uuid.UUID, categoryID int16) error {
	q := `DELETE FROM model_categories WHERE model_profile_id=$1 AND category_id=$2`
	_, err := r.pool.Exec(ctx, q, profileID, categoryID)
	return err
}

func (r *ModelRepository) GetCategories(ctx context.Context, profileID uuid.UUID) ([]domain.Category, error) {
	q := `SELECT c.id, c.name FROM categories c
		  JOIN model_categories mc ON c.id = mc.category_id
		  WHERE mc.model_profile_id = $1 ORDER BY c.name`
	rows, err := r.pool.Query(ctx, q, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

// Search

type ModelSearchFilter struct {
	City              string
	MinAge            *int
	MaxAge            *int
	CategoryID        *int16
	WillingToRelocate *bool
	Limit             int
	Offset            int
}

func (r *ModelRepository) Search(ctx context.Context, f ModelSearchFilter) ([]domain.ModelProfile, int, error) {
	args := []any{}
	conditions := []string{
		"u.status = 'active'",
		"EXISTS (SELECT 1 FROM model_photos mp WHERE mp.model_profile_id = p.id AND mp.moderation_status = 'approved')",
	}
	argIdx := 1

	if f.City != "" {
		conditions = append(conditions, fmt.Sprintf("LOWER(p.city) LIKE LOWER($%d)", argIdx))
		args = append(args, "%"+f.City+"%")
		argIdx++
	}
	if f.MinAge != nil {
		conditions = append(conditions, fmt.Sprintf("EXTRACT(YEAR FROM AGE(p.birth_date)) >= $%d", argIdx))
		args = append(args, *f.MinAge)
		argIdx++
	}
	if f.MaxAge != nil {
		conditions = append(conditions, fmt.Sprintf("EXTRACT(YEAR FROM AGE(p.birth_date)) <= $%d", argIdx))
		args = append(args, *f.MaxAge)
		argIdx++
	}
	if f.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("EXISTS (SELECT 1 FROM model_categories mc WHERE mc.model_profile_id = p.id AND mc.category_id = $%d)", argIdx))
		args = append(args, *f.CategoryID)
		argIdx++
	}
	if f.WillingToRelocate != nil {
		conditions = append(conditions, fmt.Sprintf("p.willing_to_relocate = $%d", argIdx))
		args = append(args, *f.WillingToRelocate)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM model_profiles p JOIN users u ON u.id = p.user_id %s`, where)
	var total int
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if f.Limit == 0 {
		f.Limit = 20
	}

	args = append(args, f.Limit, f.Offset)
	q := fmt.Sprintf(`SELECT p.id, p.user_id, p.first_name, p.last_name, p.birth_date, p.city, p.country,
		  p.willing_to_relocate, p.height_cm, p.weight_kg, p.chest_cm, p.waist_cm, p.hips_cm, p.shoe_size,
		  p.clothing_size, p.gender, p.hair_color, p.hair_length, p.hair_structure, p.eye_color,
		  p.clothing_size_top, p.clothing_size_bot, p.phone, p.bio, p.shoot_restrictions, p.created_at, p.updated_at,
		  COALESCE((SELECT storage_key FROM model_photos
		            WHERE model_profile_id = p.id AND is_cover = true AND moderation_status = 'approved'
		            LIMIT 1), '') AS cover_photo_storage_key
		  FROM model_profiles p JOIN users u ON u.id = p.user_id
		  %s ORDER BY p.updated_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("search models: %w", err)
	}
	defer rows.Close()

	var profiles []domain.ModelProfile
	for rows.Next() {
		var p domain.ModelProfile
		err := rows.Scan(
			&p.ID, &p.UserID, &p.FirstName, &p.LastName, &p.BirthDate,
			&p.City, &p.Country, &p.WillingToRelocate,
			&p.HeightCm, &p.WeightKg, &p.ChestCm, &p.WaistCm, &p.HipsCm, &p.ShoeSize,
			&p.ClothingSize, &p.Gender, &p.HairColor, &p.HairLength, &p.HairStructure, &p.EyeColor,
			&p.ClothingSizeTop, &p.ClothingSizeBot,
			&p.Phone, &p.Bio, &p.ShootRestrictions,
			&p.CreatedAt, &p.UpdatedAt,
			&p.CoverPhotoStorageKey,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan search result: %w", err)
		}
		profiles = append(profiles, p)
	}
	return profiles, total, rows.Err()
}

// Photos

func (r *ModelRepository) AddPhoto(ctx context.Context, photo *domain.ModelPhoto) error {
	q := `INSERT INTO model_photos (id, model_profile_id, storage_key, original_name, size_bytes,
		  mime_type, width_px, height_px, is_cover, sort_order)
		  VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.pool.Exec(ctx, q,
		photo.ID, photo.ModelProfileID, photo.StorageKey, photo.OriginalName, photo.SizeBytes,
		photo.MimeType, photo.WidthPx, photo.HeightPx, photo.IsCover, photo.SortOrder)
	return err
}

func (r *ModelRepository) GetPhoto(ctx context.Context, id uuid.UUID) (*domain.ModelPhoto, error) {
	q := `SELECT id, model_profile_id, storage_key, original_name, size_bytes, mime_type,
		  width_px, height_px, moderation_status, moderated_by, moderated_at,
		  rejection_reason, is_cover, sort_order, created_at
		  FROM model_photos WHERE id = $1`
	row := r.pool.QueryRow(ctx, q, id)
	return scanPhoto(row)
}

func (r *ModelRepository) GetPhotos(ctx context.Context, profileID uuid.UUID, approvedOnly bool) ([]domain.ModelPhoto, error) {
	q := `SELECT id, model_profile_id, storage_key, original_name, size_bytes, mime_type,
		  width_px, height_px, moderation_status, moderated_by, moderated_at,
		  rejection_reason, is_cover, sort_order, created_at
		  FROM model_photos WHERE model_profile_id = $1`
	if approvedOnly {
		q += " AND moderation_status = 'approved'"
	}
	q += " ORDER BY sort_order ASC, created_at ASC"

	rows, err := r.pool.Query(ctx, q, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []domain.ModelPhoto
	for rows.Next() {
		p, err := scanPhoto(rows)
		if err != nil {
			return nil, err
		}
		photos = append(photos, *p)
	}
	return photos, rows.Err()
}

func (r *ModelRepository) UpdatePhoto(ctx context.Context, photo *domain.ModelPhoto) error {
	q := `UPDATE model_photos SET is_cover=$1, sort_order=$2 WHERE id=$3`
	_, err := r.pool.Exec(ctx, q, photo.IsCover, photo.SortOrder, photo.ID)
	return err
}

func (r *ModelRepository) DeletePhoto(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM model_photos WHERE id=$1`
	_, err := r.pool.Exec(ctx, q, id)
	return err
}

func (r *ModelRepository) ModeratePhoto(ctx context.Context, photoID uuid.UUID, status domain.ModerationStatus, moderatorID uuid.UUID, reason string) error {
	now := time.Now()
	q := `UPDATE model_photos SET moderation_status=$1, moderated_by=$2, moderated_at=$3, rejection_reason=$4 WHERE id=$5`
	_, err := r.pool.Exec(ctx, q, status, moderatorID, now, reason, photoID)
	return err
}

func (r *ModelRepository) GetPendingPhotos(ctx context.Context, limit, offset int) ([]domain.ModelPhoto, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM model_photos WHERE moderation_status='pending'`).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := `SELECT id, model_profile_id, storage_key, original_name, size_bytes, mime_type,
		  width_px, height_px, moderation_status, moderated_by, moderated_at,
		  rejection_reason, is_cover, sort_order, created_at
		  FROM model_photos WHERE moderation_status='pending'
		  ORDER BY created_at ASC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var photos []domain.ModelPhoto
	for rows.Next() {
		p, err := scanPhoto(rows)
		if err != nil {
			return nil, 0, err
		}
		photos = append(photos, *p)
	}
	return photos, total, rows.Err()
}

func (r *ModelRepository) SetPhotoTags(ctx context.Context, photoID uuid.UUID, tagIDs []int16) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM model_photo_tags WHERE photo_id=$1`, photoID)
	if err != nil {
		return err
	}
	for _, tid := range tagIDs {
		_, err := r.pool.Exec(ctx, `INSERT INTO model_photo_tags (photo_id, photo_tag_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, photoID, tid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ModelRepository) CountPhotos(ctx context.Context, profileID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM model_photos WHERE model_profile_id=$1`, profileID).Scan(&count)
	return count, err
}

// Reference data

func (r *ModelRepository) GetAllCategories(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *ModelRepository) GetAllPhotoTags(ctx context.Context) ([]domain.PhotoTag, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name FROM photo_tags ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []domain.PhotoTag
	for rows.Next() {
		var t domain.PhotoTag
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

// Scanners

func scanModelProfile(row scanner) (*domain.ModelProfile, error) {
	var p domain.ModelProfile
	err := row.Scan(
		&p.ID, &p.UserID, &p.FirstName, &p.LastName, &p.BirthDate,
		&p.City, &p.Country, &p.WillingToRelocate,
		&p.HeightCm, &p.WeightKg, &p.ChestCm, &p.WaistCm, &p.HipsCm, &p.ShoeSize,
		&p.ClothingSize, &p.Gender, &p.HairColor, &p.HairLength, &p.HairStructure, &p.EyeColor,
		&p.ClothingSizeTop, &p.ClothingSizeBot,
		&p.Phone, &p.Bio, &p.ShootRestrictions,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan model profile: %w", err)
	}
	return &p, nil
}

func scanPhoto(row scanner) (*domain.ModelPhoto, error) {
	var p domain.ModelPhoto
	err := row.Scan(
		&p.ID, &p.ModelProfileID, &p.StorageKey, &p.OriginalName, &p.SizeBytes,
		&p.MimeType, &p.WidthPx, &p.HeightPx,
		&p.ModerationStatus, &p.ModeratedBy, &p.ModeratedAt,
		&p.RejectionReason, &p.IsCover, &p.SortOrder, &p.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan photo: %w", err)
	}
	return &p, nil
}
