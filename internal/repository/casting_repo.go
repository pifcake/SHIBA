package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"SHIBA/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CastingRepository struct {
	pool *pgxpool.Pool
}

func NewCastingRepository(pool *pgxpool.Pool) *CastingRepository {
	return &CastingRepository{pool: pool}
}

func (r *CastingRepository) Create(ctx context.Context, c *domain.Casting) error {
	q := `INSERT INTO castings (id, agency_profile_id, title, description, city, casting_date, category_id, status)
		  VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.pool.Exec(ctx, q, c.ID, c.AgencyProfileID, c.Title, c.Description, c.City, c.CastingDate, c.CategoryID, c.Status)
	return err
}

func (r *CastingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Casting, error) {
	q := `SELECT id, agency_profile_id, title, description, city, casting_date, category_id, status, created_at, updated_at
		  FROM castings WHERE id=$1`
	row := r.pool.QueryRow(ctx, q, id)
	return scanCasting(row)
}

func (r *CastingRepository) Update(ctx context.Context, c *domain.Casting) error {
	q := `UPDATE castings SET title=$1, description=$2, city=$3, casting_date=$4,
		  category_id=$5, status=$6, updated_at=NOW() WHERE id=$7`
	ct, err := r.pool.Exec(ctx, q, c.Title, c.Description, c.City, c.CastingDate, c.CategoryID, c.Status, c.ID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CastingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	q := `UPDATE castings SET status='cancelled', updated_at=NOW() WHERE id=$1`
	_, err := r.pool.Exec(ctx, q, id)
	return err
}

type CastingFilter struct {
	Status     string
	City       string
	CategoryID *int16
	AgencyID   *uuid.UUID
	Limit      int
	Offset     int
}

func (r *CastingRepository) List(ctx context.Context, f CastingFilter) ([]domain.Casting, int, error) {
	conditions := []string{"1=1"}
	args := []any{}
	argIdx := 1

	if f.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status=$%d", argIdx))
		args = append(args, f.Status)
		argIdx++
	}
	if f.City != "" {
		conditions = append(conditions, fmt.Sprintf("LOWER(city) LIKE LOWER($%d)", argIdx))
		args = append(args, "%"+f.City+"%")
		argIdx++
	}
	if f.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("category_id=$%d", argIdx))
		args = append(args, *f.CategoryID)
		argIdx++
	}
	if f.AgencyID != nil {
		conditions = append(conditions, fmt.Sprintf("agency_profile_id=$%d", argIdx))
		args = append(args, *f.AgencyID)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM castings %s", where), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if f.Limit == 0 {
		f.Limit = 20
	}
	args = append(args, f.Limit, f.Offset)
	q := fmt.Sprintf(`SELECT id, agency_profile_id, title, description, city, casting_date, category_id, status, created_at, updated_at
		  FROM castings %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var castings []domain.Casting
	for rows.Next() {
		c, err := scanCasting(rows)
		if err != nil {
			return nil, 0, err
		}
		castings = append(castings, *c)
	}
	return castings, total, rows.Err()
}

// Applications

func (r *CastingRepository) CreateApplication(ctx context.Context, a *domain.Application) error {
	q := `INSERT INTO applications (id, casting_id, model_profile_id, status, model_message)
		  VALUES ($1,$2,$3,$4,$5)`
	_, err := r.pool.Exec(ctx, q, a.ID, a.CastingID, a.ModelProfileID, a.Status, a.ModelMessage)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrConflict
		}
		return err
	}
	return nil
}

func (r *CastingRepository) GetApplication(ctx context.Context, id uuid.UUID) (*domain.Application, error) {
	q := `SELECT id, casting_id, model_profile_id, status, model_message, agency_response, created_at, updated_at
		  FROM applications WHERE id=$1`
	row := r.pool.QueryRow(ctx, q, id)
	return scanApplication(row)
}

func (r *CastingRepository) GetApplicationByModelAndCasting(ctx context.Context, castingID, modelID uuid.UUID) (*domain.Application, error) {
	q := `SELECT id, casting_id, model_profile_id, status, model_message, agency_response, created_at, updated_at
		  FROM applications WHERE casting_id=$1 AND model_profile_id=$2`
	row := r.pool.QueryRow(ctx, q, castingID, modelID)
	return scanApplication(row)
}

func (r *CastingRepository) GetApplicationsByCasting(ctx context.Context, castingID uuid.UUID, limit, offset int) ([]domain.Application, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM applications WHERE casting_id=$1`, castingID).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := `SELECT id, casting_id, model_profile_id, status, model_message, agency_response, created_at, updated_at
		  FROM applications WHERE casting_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, q, castingID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var apps []domain.Application
	for rows.Next() {
		a, err := scanApplication(rows)
		if err != nil {
			return nil, 0, err
		}
		apps = append(apps, *a)
	}
	return apps, total, rows.Err()
}

func (r *CastingRepository) GetApplicationsByModel(ctx context.Context, modelID uuid.UUID, limit, offset int) ([]domain.Application, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM applications WHERE model_profile_id=$1`, modelID).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := `SELECT id, casting_id, model_profile_id, status, model_message, agency_response, created_at, updated_at
		  FROM applications WHERE model_profile_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, q, modelID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var apps []domain.Application
	for rows.Next() {
		a, err := scanApplication(rows)
		if err != nil {
			return nil, 0, err
		}
		apps = append(apps, *a)
	}
	return apps, total, rows.Err()
}

func (r *CastingRepository) UpdateApplicationStatus(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus, response string) error {
	q := `UPDATE applications SET status=$1, agency_response=$2, updated_at=NOW() WHERE id=$3`
	ct, err := r.pool.Exec(ctx, q, status, response, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CastingRepository) DeleteApplication(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM applications WHERE id=$1`
	_, err := r.pool.Exec(ctx, q, id)
	return err
}

// Invitations

func (r *CastingRepository) CreateInvitation(ctx context.Context, inv *domain.Invitation) error {
	q := `INSERT INTO invitations (id, agency_profile_id, model_profile_id, casting_id, message, status)
		  VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := r.pool.Exec(ctx, q, inv.ID, inv.AgencyProfileID, inv.ModelProfileID, inv.CastingID, inv.Message, inv.Status)
	return err
}

func (r *CastingRepository) GetInvitation(ctx context.Context, id uuid.UUID) (*domain.Invitation, error) {
	q := `SELECT id, agency_profile_id, model_profile_id, casting_id, message, status, created_at, updated_at
		  FROM invitations WHERE id=$1`
	row := r.pool.QueryRow(ctx, q, id)
	return scanInvitation(row)
}

func (r *CastingRepository) GetInvitationsByModel(ctx context.Context, modelID uuid.UUID, limit, offset int) ([]domain.Invitation, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM invitations WHERE model_profile_id=$1`, modelID).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := `SELECT id, agency_profile_id, model_profile_id, casting_id, message, status, created_at, updated_at
		  FROM invitations WHERE model_profile_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, q, modelID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var invs []domain.Invitation
	for rows.Next() {
		inv, err := scanInvitation(rows)
		if err != nil {
			return nil, 0, err
		}
		invs = append(invs, *inv)
	}
	return invs, total, rows.Err()
}

func (r *CastingRepository) GetInvitationsByAgency(ctx context.Context, agencyID uuid.UUID, limit, offset int) ([]domain.Invitation, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM invitations WHERE agency_profile_id=$1`, agencyID).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := `SELECT id, agency_profile_id, model_profile_id, casting_id, message, status, created_at, updated_at
		  FROM invitations WHERE agency_profile_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, q, agencyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var invs []domain.Invitation
	for rows.Next() {
		inv, err := scanInvitation(rows)
		if err != nil {
			return nil, 0, err
		}
		invs = append(invs, *inv)
	}
	return invs, total, rows.Err()
}

func (r *CastingRepository) UpdateInvitationStatus(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus) error {
	q := `UPDATE invitations SET status=$1, updated_at=NOW() WHERE id=$2`
	ct, err := r.pool.Exec(ctx, q, status, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Scanners

func scanCasting(row scanner) (*domain.Casting, error) {
	var c domain.Casting
	err := row.Scan(
		&c.ID, &c.AgencyProfileID, &c.Title, &c.Description, &c.City,
		&c.CastingDate, &c.CategoryID, &c.Status, &c.CreatedAt, &c.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan casting: %w", err)
	}
	return &c, nil
}

func scanApplication(row scanner) (*domain.Application, error) {
	var a domain.Application
	err := row.Scan(
		&a.ID, &a.CastingID, &a.ModelProfileID, &a.Status,
		&a.ModelMessage, &a.AgencyResponse, &a.CreatedAt, &a.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan application: %w", err)
	}
	return &a, nil
}

func scanInvitation(row scanner) (*domain.Invitation, error) {
	var inv domain.Invitation
	err := row.Scan(
		&inv.ID, &inv.AgencyProfileID, &inv.ModelProfileID, &inv.CastingID,
		&inv.Message, &inv.Status, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan invitation: %w", err)
	}
	return &inv, nil
}
